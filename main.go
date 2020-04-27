package main

import (
	"context"
	"crypto/sha256"
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

var (
	version = "0.0.0" // goreleaser sets this value
)

const (
	stateFiring     = "firing"
	stateResolved   = "resolved"
	timeout         = 10 * time.Second
	dfltURL         = "http://localhost:9090/api/v1/alerts"
	alertNamePrefix = "testalert"
)

type alert struct {
	State       string `json:"state"`
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

	l := labels{
		"alertname": name(),
		"instance":  name() + ".example.net",
	}
	user := os.Getenv("USER")
	summary := flag.String("summary", "This is a test alert", "The summary for the alert.")
	url := flag.String("url", dfltURL, "The prometheus URL.")
	v := flag.Bool("version", false, "Show version information.")
	flag.Parse()

	if *v {
		fmt.Println("version:", version)
		return
	}

	if user != "" {
		l["user"] = user
	}

	flag.Var(&l, "labels", "The labels to use for the alert.")

	c := client{
		url: *url,
		Client: &http.Client{
			Timeout: timeout,
		},
	}

	a := alert{
		State:  stateFiring,
		Labels: l,
		Annotations: annotations{
			Summary: *summary,
		},
	}

	if err := c.post(a); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("alert '%s' is %s\n", a.Annotations.Summary, a.State)
	fmt.Println("ctrl+c to resolve alert")
	<-ctx.Done()

	a.State = stateResolved

	fmt.Printf("alert '%s' is %s\n", a.Annotations.Summary, a.State)

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

func name() string {
	h := sha256.New()
	h.Write([]byte(time.Now().String()))
	sha := fmt.Sprintf("%x", h.Sum(nil))

	return alertNamePrefix + "-" + sha[:8]
}
