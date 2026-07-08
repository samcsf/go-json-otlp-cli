package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"

	"go.opentelemetry.io/contrib/bridges/otelslog"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	js := map[string]interface{}{}
	for scanner.Scan() {
		if err := json.Unmarshal(scanner.Bytes(), &js); err != nil {
			panic(err)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "read error:", err)
	}
	ctx := context.Background()
	exp, err := otlploghttp.New(ctx, otlploghttp.WithInsecure())
	if err != nil {
		panic(err)
	}

	processor := log.NewBatchProcessor(exp)
	provider := log.NewLoggerProvider(
		log.WithProcessor(processor),
		log.WithResource(resource.NewSchemaless(attribute.KeyValue{
			Key:   "service.name",
			Value: attribute.StringValue("claude-code"),
		})),
	)
	defer func() {
		if err := provider.Shutdown(ctx); err != nil {
			panic(err)
		}
	}()

	global.SetLoggerProvider(provider)
	lg := otelslog.NewLogger("default")

	args := make([]any, 0, len(js)*2)
	for k, v := range js {
		args = append(args, k, v)
	}
	lg.InfoContext(ctx, "Received event", args...)
}
