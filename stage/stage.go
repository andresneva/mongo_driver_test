package stage

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/andresneva/mongo_driver_test/repositories"
	"github.com/andresneva/mongo_driver_test/stats"
)

var timeouts int64

//Config struct
type Config struct {
	WorkersCount     uint
	WorkersToAdd     uint
	IncrementLoad    uint
	ProducersCount   uint
	MsgBySec         uint
	TimeToSleepSecs  uint
	TimeToFinishSecs uint
	QueryTimeoutMs   uint
	BatchSize        int32
	CollectionSize   int
	DocumentSize     int
}

//Stage struct
type Stage struct {
	dbConfig    repositories.MongoDBConfiguration
	stageConfig Config
}

//New stage
func New(
	dbConfig repositories.MongoDBConfiguration,
	stageConfig Config) *Stage {
	return &Stage{
		dbConfig:    dbConfig,
		stageConfig: stageConfig,
	}
}

//Run starts the test
func (s *Stage) Run(id string) {

	timeouts = 0

	statsMonitor := stats.NewPoolStats()

	config := &repositories.MongoDBConfiguration{
		DbName:         s.dbConfig.DbName,
		CollectionName: s.dbConfig.CollectionName,
		ConnString:     s.dbConfig.ConnString,
		MinPool:        s.dbConfig.MinPool,
		MaxPool:        s.dbConfig.MaxPool,
		IdleTimeout:    s.dbConfig.IdleTimeout,
		SocketTimeout:  s.dbConfig.SocketTimeout,
	}
	repo, err := repositories.NewMongodbRepository(config, statsMonitor.MonitorFunc)
	if err != nil {
		logrus.Fatal(err)
	}

	storeIds, err := ensureData(repo, s.stageConfig.CollectionSize, s.stageConfig.DocumentSize)
	if err != nil {
		return
	}

	repo.SetValidIds(storeIds)

	eventChannel := make(chan struct{}, 1000)

	wgP := &sync.WaitGroup{}

	producers := addProducers(int(s.stageConfig.ProducersCount), eventChannel, int(s.stageConfig.MsgBySec), wgP)

	workers := addWorkers(int(s.stageConfig.WorkersCount), repo, eventChannel, s.stageConfig.QueryTimeoutMs, s.stageConfig.BatchSize)

	intLoad := int(s.stageConfig.IncrementLoad)
	intTimeToSleep := int(s.stageConfig.TimeToSleepSecs)
	for n := 0; n < intLoad; n++ {
		logrus.Printf("Waiting %d seconds to add %d workers. Current count: %d",
			s.stageConfig.TimeToSleepSecs, s.stageConfig.WorkersToAdd, len(workers))
		for i := 0; i < intTimeToSleep; i++ {
			logrus.WithField("executed", repo.QueryCount()).Infof("%v", statsMonitor)
			time.Sleep(1 * time.Second)
		}
		workers = append(workers, addWorkers(int(s.stageConfig.WorkersToAdd), repo, eventChannel, s.stageConfig.QueryTimeoutMs, s.stageConfig.BatchSize)...)
		logrus.Printf("%d workers added. Using %d in total", s.stageConfig.WorkersToAdd, len(workers))
	}

	logrus.Printf("Waiting %d seconds to finish", s.stageConfig.TimeToFinishSecs)
	intTimeToFinish := int(s.stageConfig.TimeToFinishSecs)
	for i := 0; i < intTimeToFinish; i++ {
		logrus.WithField("executed", repo.QueryCount()).Infof("%+v", statsMonitor)
		time.Sleep(1 * time.Second)
	}

	for _, producer := range producers {
		producer.stop()
	}
	wgP.Wait()
	logrus.Println("Producers stopped.")

	for len(eventChannel) > 0 {
		logrus.WithField("executed", repo.QueryCount()).Infof("%+v", statsMonitor)
		time.Sleep(1 * time.Second)
	}

	repo.Close()

	time.Sleep(1 * time.Second)

	logrus.Printf("")
	logrus.Printf("--------------------------------------------------------------------------------------------------------------")
	logrus.Printf("Final stats: %+v", statsMonitor)
	logrus.Printf("--------------------------------------------------------------------------------------------------------------")
	logrus.Printf("")

	logrus.Printf("************************************")
	logrus.Printf("Total query count: %d", repo.QueryCount())
	logrus.Printf("Total query timeouts: %d", timeouts)
	logrus.Printf("Timeout percentage: %s", TimeoutPercentage(repo.QueryCount()))
	logrus.Printf("************************************")

}

//TimeoutPercentage calculates the percentag and returns a string
func TimeoutPercentage(queryCount int64) string {
	timeoutPercentage := 100 * float64(timeouts) / float64(queryCount)

	timeoutsString := fmt.Sprintf("%.2f", timeoutPercentage)

	timeoutsString = timeoutsString + "%"

	return timeoutsString
}

func addWorkers(
	workersCount int,
	repo repositories.TestRepository,
	evChan chan struct{},
	timeout uint,
	batchSize int32,
) []*consumer {
	var consumers []*consumer
	for i := 0; i < workersCount; i++ {
		consumer := &consumer{
			repository:   repo,
			eventChannel: evChan,
			timeout:      timeout,
			batchSize:    batchSize,
		}
		consumers = append(consumers, consumer)
		go consumer.start()
	}
	return consumers
}

func addProducers(producersCount int, eventChannel chan struct{}, msgBySec int, wg *sync.WaitGroup) []*producer {
	var producers []*producer

	wg.Add(producersCount)

	for i := 0; i < producersCount; i++ {
		producer := &producer{
			eventChannel: eventChannel,
			wg:           wg,
		}
		producers = append(producers, producer)

		go producer.start(time.Duration(1000/msgBySec) * time.Millisecond)
	}

	return producers
}

type producer struct {
	eventChannel chan<- struct{}
	tm           *time.Ticker
	wg           *sync.WaitGroup
}

func (p *producer) start(sendEvery time.Duration) {
	p.tm = time.NewTicker(sendEvery)
	for range p.tm.C {
		p.eventChannel <- struct{}{}
	}
}

func (p *producer) stop() {
	p.tm.Stop()
	time.Sleep(100 * time.Millisecond)
	p.wg.Done()
}

type consumer struct {
	repository   repositories.TestRepository
	timeout      uint
	batchSize    int32
	eventChannel <-chan struct{}
}

func (c *consumer) start() {

	for range c.eventChannel {
		size := rand.Intn(400-100) + 100 //pseudo random it's ok

		_, executionTime, err := c.repository.GetStores(uint(size), c.timeout, c.batchSize)
		if err != nil {
			timeouts = timeouts + 1
			logrus.WithField("Execution time", executionTime).Errorf("%+v", err)
		}
	}
}

func ensureData(repository repositories.TestRepository, collectionSize int, documentSize int) ([]string, error) {

	logrus.Info("Creating data for the test...")

	count, err := repository.Count()
	if err != nil {
		return nil, err
	}

	if count > 0 {
		repository.Clear()
	}

	var storeIds []string
	var data []repositories.Store
	for i := 0; i < collectionSize; i++ {
		storeID := GenerateID()
		storeIds = append(storeIds, storeID)
		name := "name: " + strconv.Itoa(i)

		var content string

		content = GenerateString(documentSize)

		//objectSize := len(storeID) + len(name) + len(content)

		//logrus.Info("Document size in bytes: ", objectSize)

		data = append(data, repositories.Store{
			StoreId:   storeID,
			Name:      name,
			HugeValue: content,
		})

		logrus.Infof("Creating document %d of %d", i+1, collectionSize)
	}
	logrus.Infof("%d Documents created.", collectionSize)
	logrus.Infof("Inserting data into the database...")
	err = repository.Insert(data)
	logrus.Infof("Data inserted.")

	return storeIds, err
}
