package otexample

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	tracer      = otel.Tracer("rolldice")
	meter       = otel.Meter("rolldice")
	rollCounter metric.Int64Counter
)

func init() {
	var err error
	rollCounter, err = meter.Int64Counter("dice.rolls",
		metric.WithDescription("The number of rolls by roll value"),
		metric.WithUnit("{roll}"))
	if err != nil {
		panic(err)
	}
}

func RollDice(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "roll")
	defer span.End()

	roll := 1 + rand.Intn(6)

	// add a custom atribute to the span and counter
	rollValueAttr := attribute.Int("roll.value", roll)
	span.SetAttributes(rollValueAttr)
	rollCounter.Add(ctx, 1, metric.WithAttributes(rollValueAttr))

	if _, err := fmt.Fprintf(w, "%d\n", roll); err != nil {
		log.Print(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}
