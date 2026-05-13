package sonar

import (
	"testing"

	"rover-kit/pkg/uart"
)

func TestSampleToReadingsConvertsMillimeters(t *testing.T) {
	readings := sampleToReadings(uart.SampleV1{
		DistanceUnit: uart.DistanceUnitMillimeters,
		Readings:     []uint16{250, 0xFFFF, 1234},
	})

	if len(readings) != 2 {
		t.Fatalf("expected 2 valid readings, got %d", len(readings))
	}
	if readings[0].DistanceCM != 25.0 {
		t.Fatalf("expected first reading to be 25.0cm, got %f", readings[0].DistanceCM)
	}
	if readings[1].DistanceCM != 123.4 {
		t.Fatalf("expected second reading to be 123.4cm, got %f", readings[1].DistanceCM)
	}
}

func TestSampleToReadingsKeepsCentimeters(t *testing.T) {
	readings := sampleToReadings(uart.SampleV1{
		DistanceUnit: uart.DistanceUnitCentimeters,
		Readings:     []uint16{42},
	})

	if len(readings) != 1 {
		t.Fatalf("expected 1 reading, got %d", len(readings))
	}
	if readings[0].DistanceCM != 42 {
		t.Fatalf("expected reading to be 42cm, got %f", readings[0].DistanceCM)
	}
}
