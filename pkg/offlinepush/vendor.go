package offlinepush

import (
	"context"
	"log"
	"os"
)

type Vendor struct {
	enabled bool
}

func NewVendor() *Vendor {
	return &Vendor{enabled: os.Getenv("DEGRADE_OFFLINE_PUSH") != "1"}
}

func (v *Vendor) Send(ctx context.Context, userID int64, title, body string) error {
	if !v.enabled {
		log.Printf("offline push degraded, skip uid=%d", userID)
		return nil
	}
	log.Printf("[APNs/FCM stub] uid=%d title=%q body=%q", userID, title, body)
	return nil
}
