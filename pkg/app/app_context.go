package app

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"time"

	"github.com/go-co-op/gocron/v2"
	amqp "github.com/rabbitmq/amqp091-go"
	"gitlab.com/Hamed1984/stats/pkg/conf"
	"gitlab.com/Hamed1984/stats/pkg/db"
	"gitlab.com/Hamed1984/stats/pkg/logging"
	"gitlab.com/Hamed1984/stats/pkg/mq"
	"gitlab.com/Hamed1984/stats/pkg/provider"
	"go.uber.org/zap"
)

type Application struct {
	config      *conf.Configuration
	ordersRepo  db.OrderRepo
	historyRepo db.HistoryRepo
	statsMQ     mq.StatusPublisherReceiver
	doneCh      chan struct{}
	tokensCh    chan struct{}
	statusCh    chan *mq.StatusMsg
	statusErrCh chan *mq.StatusPubError
	scheduler   gocron.Scheduler
}

func NewApplication(conf *conf.Configuration) (*Application, error) {
	done := make(chan struct{})
	tokens := make(chan struct{}, conf.ParallelismLevel)

	dbConn, err := db.NewMysqlConnection(conf)
	if err != nil {
		return nil, err
	}

	ordRepo := db.NewOrderRepoImpl(dbConn)
	histRepo := db.NewHistoryRepoImpl(dbConn)

	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	sMQ, err := mq.NewStatusMQ(conf, done)
	if err != nil {
		return nil, err
	}

	sMQ.StartConsume(func(d amqp.Delivery, done <-chan struct{}) error {
		select {
		case <-done:
			return nil
		default:
			var msg mq.StatusMsg
			err := json.Unmarshal(d.Body, &msg)
			if err != nil {
				logging.GetLogger(conf).Error(err.Error())
				d.Ack(false)
				return nil
			}
			ord := &db.Order{
				OrderID:       msg.OrderID,
				CurrentStatus: msg.Status,
				StatusEP:      msg.StatusEP,
			}
			ctx, cancelF := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancelF()
			_, err = ordRepo.CreateOrder(ctx, ord)
			if err != nil {
				logging.GetLogger(conf).Error(err.Error())
				d.Nack(false, false)
				return nil
			}
			d.Ack(false)

		}
		return nil
	})
	statCh := make(chan *mq.StatusMsg)
	statErrCh := make(chan *mq.StatusPubError)

	go func() {

		for x := range statErrCh {
			logging.GetLogger(conf).Error(x.Err.Error())
			logging.GetLogger(conf).Info("retry publising message", zap.Uint64("order_id", x.Msg.OrderID))
			statCh <- x.Msg
		}

	}()
	sMQ.StartPublisher(statCh, statErrCh)

	ret := &Application{
		config:      conf,
		ordersRepo:  ordRepo,
		historyRepo: histRepo,
		statsMQ:     sMQ,
		doneCh:      done,
		tokensCh:    tokens,
		statusCh:    statCh,
		statusErrCh: statErrCh,
		scheduler:   scheduler,
	}
	return ret, nil
}

func (a *Application) Start() error {
	var du time.Duration
	du = 24 * time.Hour
	if a.config.Development {
		du = 40 * time.Second
	}
	j, err := a.scheduler.NewJob(
		gocron.DurationJob(du),
		gocron.NewTask(func() {
			logging.GetLogger(a.config).Info("job started")
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			data, err := a.ordersRepo.GetPickedupNoneDeliveredOrders(ctx)
			if err != nil {
				logging.GetLogger(a.config).Error(err.Error())
				return
			}

			for _, s := range data {
				a.tokensCh <- struct{}{}
				go func(o db.Order) {
					defer func() {
						<-a.tokensCh
					}()
					logging.GetLogger(a.config).Info("getting order status for order", zap.Any("order_id", o.OrderID))
					stats, err := provider.GetOrderStatusFromProvider(o.StatusEP, o.OrderID)
					if err != nil {
						logging.GetLogger(a.config).Error(err.Error())
						return
					}
					var hists []db.History
					for _, x := range stats.Data {
						hists = append(hists, db.History{
							Status:    x.Status,
							OrderID:   o.OrderID,
							CreatedAt: x.CreatedAt,
						})
					}
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					err = a.historyRepo.CreateHistories(ctx, hists)
					if err != nil {
						logging.GetLogger(a.config).Error(err.Error())
					}
					for _, x := range stats.Data {
						go func(p provider.OrderStatus) {
							msg := &mq.StatusMsg{
								OrderID:   o.OrderID,
								Status:    p.Status,
								StatusEP:  o.StatusEP,
								CreatedAt: p.CreatedAt,
							}
							select {
							case <-a.doneCh:
								return
							case a.statusCh <- msg:
								return
							}
						}(x)
					}

				}(s)
			}

		}),
	)
	if err != nil {
		return err
	}
	logging.GetLogger(a.config).Info("job id", zap.Any("job_id", j.ID()))
	a.scheduler.Start()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
	close(a.doneCh)
	return a.scheduler.Shutdown()
}
