package worker

import (
	"AIGateway/internal/aiclient"
	"context"
	"log"
	"sync"
)

type Job struct {
	EventID int
	Payload string
}
type Result struct {
	EventID int
	Content string
	Model   string
}

type Enrichments interface {
	Save(ctx context.Context, EventId int, resp string, model string) error
}

// TODO: Добавить реальный репо для сохранения в БД
type Pipeline struct {
	in          chan Job
	aiClient    *aiclient.Client
	wg          *sync.WaitGroup
	repo        Enrichments
	workerCount int
	buffer      int
}

func NewPipeline(in chan Job, aiClient *aiclient.Client, wg *sync.WaitGroup, repo Enrichments, workerCount int, buffer int) *Pipeline {
	return &Pipeline{in: in, aiClient: aiClient, wg: wg, repo: repo, workerCount: workerCount, buffer: buffer}
}

func (p *Pipeline) Producer(ctx context.Context, EventId int, payload string) error {
	//Собираем структуру Job,
	//записываем ее в общий канал
	job := Job{EventID: EventId, Payload: payload}
	select {
	case p.in <- job:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}

}

func (p *Pipeline) Worker(ctx context.Context) <-chan Result {
	//Создаем выходной канал
	out := make(chan Result)

	go func() {
		//Итерируемся по общему каналу,
		//и вызываем Fetch для обработки payload
		for v := range p.in {
			resp, err := p.aiClient.Fetch(ctx, v.Payload)
			if err != nil {
				//Полностью не убиваем функцию,
				//а только пропускаем
				log.Println("fetch err:", err)
				continue
			}
			select {
			//В случае успешной записи в канал,
			//пропускам второй кейс и переходим обратно в цикл
			case out <- Result{EventID: v.EventID, Content: resp.OutputText(), Model: resp.Model}:
				continue
			case <-ctx.Done():
				return
			}
		}
		close(out)
	}()
	return out
}

func (p *Pipeline) Merge(channels ...<-chan Result) <-chan Result {
	out := make(chan Result)

	//Функция для чтения из 1 конкретного канала
	output := func(ch <-chan Result) {
		//Читаем из канала и записываем в out
		for n := range ch {
			out <- n
		}
		p.wg.Done()
	}
	//Количество каналов переданных в конкретный вызов
	p.wg.Add(len(channels))
	//итерируемся по каналам,
	//на каждой итерации вызываем функцию output
	for _, ch := range channels {
		go output(ch)
	}
	//Вторая горутина для ожидания выполнения и закрытия канала
	go func() {
		p.wg.Wait()
		close(out)
	}()
	return out
}

func (p *Pipeline) Start(ctx context.Context) {
	var channels []<-chan Result

	for i := 0; i < p.workerCount; i++ {
		channels = append(channels, p.Worker(ctx))
	}

	merged := p.Merge(channels...)

	//TODO: добавить wg для реализации Graceful Shutdown
	go func() {
		for v := range merged {
			err := p.repo.Save(ctx, v.EventID, v.Content, v.Model)
			if err != nil {
				log.Println("save err:", err)
			}
		}
	}()
}
