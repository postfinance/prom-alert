package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	statusFiring   = "firing"
	statusResolved = "resolved"
	timeout        = 10 * time.Second
	dfltURL        = "http://localhost:9090/api/v1/alerts"
)

type alert struct {
	Status      string `json:"status"`
	Labels      labels
	Annotations annotations `json:"annotations"`
}

type annotations struct {
	Summary string `json:"summary"`
}

type labels map[string]string

func (l labels) String() string {
	fields := []string{}
	for k, v := range l {
		fields = append(fields, fmt.Sprintf("%s=%s", k, v))
	}

	return strings.Join(fields, ",")
}

func (l labels) Set(s string) error {
	if s == "" {
		return nil
	}

	fields := strings.Split(s, ",")

	for _, field := range fields {
		d := strings.Split(field, "=")
		if len(d) != 2 {
			return fmt.Errorf("unsupported labels format for %s", fields)
		}

		l[d[0]] = d[1]
	}

	return nil
}

func main() {
	ctx := contextWithSignal(context.Background(), nil, syscall.SIGINT, syscall.SIGTERM)

	l := labels{}
	summary := flag.String("summary", "Latency is high", "The summary for the alert.")
	url := flag.String("url", dfltURL, "The prometheus URL.")

	flag.Var(&l, "labels", "The labels to use for the alert (for example: service=my-service,team=my-team).")
	flag.Parse()

	c := client{
		url: *url,
		Client: &http.Client{
			Timeout: timeout,
		},
	}

	a := alert{
		Status: statusFiring,
		Labels: l,
		Annotations: annotations{
			Summary: *summary,
		},
	}

	if err := c.post(a); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("alert '%s' is %s\n", a.Annotations.Summary, a.Status)
	fmt.Println("ctrl+c to resolve alert")
	<-ctx.Done()

	a.Status = statusResolved

	fmt.Printf("alert '%s' is %s\n", a.Annotations.Summary, a.Status)

	if err := c.post(a); err != nil {
		log.Fatal(err)
	}
}

func contextWithSignal(ctx context.Context, f func(s os.Signal), s ...os.Signal) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, s...)

		defer signal.Stop(c)

		select {
		case <-ctx.Done():
		case sig := <-c:
			if f != nil {
				f(sig)
			}

			cancel()
		}
	}()

	return ctx
}
