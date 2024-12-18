package proc

import "golang.org/x/sync/semaphore"

type RequestQueue semaphore.Weighted

func NewRequestQueue() *RequestQueue {
	sem := semaphore.NewWeighted(1)
	return (*RequestQueue)(sem)
}
