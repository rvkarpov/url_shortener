package storage

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/lib/pq"
	"github.com/rvkarpov/url_shortener/internal/config"
)

type Task struct {
	userID string
	urls   []string
}

type DeleteCmd struct {
	state     *DBState
	cfg       *config.Config
	inputChan chan Task
	mu        sync.Mutex
	buffers   map[string][]string
}

func (cmd *DeleteCmd) Append(userID string, urls []string) {
	cmd.inputChan <- Task{userID: userID, urls: urls}
}

func (cmd *DeleteCmd) Finalize() {
	close(cmd.inputChan)

	cmd.mu.Lock()
	for userID, urls := range cmd.buffers {
		cmd.flush(userID, urls)
	}
	cmd.mu.Unlock()
}

func (cmd *DeleteCmd) RunAsync() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cmd.mu.Lock()
			for userID, urls := range cmd.buffers {
				cmd.flush(userID, urls)
			}
			cmd.mu.Unlock()
		case task, ok := <-cmd.inputChan:
			if !ok {
				return
			}

			cmd.mu.Lock()
			urls := cmd.buffers[task.userID]
			urls = append(urls, task.urls...)

			var limit = 100
			if len(urls) >= limit {
				cmd.flush(task.userID, urls)
			} else {
				cmd.buffers[task.userID] = urls
			}
			cmd.mu.Unlock()
		}
	}
}

func (cmd *DeleteCmd) flush(userID string, urls []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := fmt.Sprintf(
		`UPDATE %s SET deletedFlag = TRUE WHERE userID = $1 AND shortUrl = ANY($2);`,
		pq.QuoteIdentifier(cmd.cfg.TableName),
	)

	_, err := cmd.state.DB.ExecContext(ctx, query, userID, pq.Array(cmd.buffers[userID]))
	if err != nil {
		log.Printf("Failed to mark URLs as deleted: %v", err)
	}

	delete(cmd.buffers, userID)
}

func NewDeleteCmd(state *DBState, cfg *config.Config) *DeleteCmd {
	inputChan := make(chan Task, 100)
	buffers := make(map[string][]string)
	cmd := &DeleteCmd{state: state, cfg: cfg, buffers: buffers, inputChan: inputChan}
	go cmd.RunAsync()

	return cmd
}
