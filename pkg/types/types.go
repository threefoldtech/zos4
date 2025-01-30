package types

import (
	"time"

	"github.com/lib/pq"
)

type NodeRegistrationRequest struct {
	TwinID       uint64      `json:"twin_id" binding:"required,min=1"`
	FarmID       uint64      `json:"farm_id" binding:"required,min=1"`
	Resources    Resources   `json:"resources" binding:"required"`
	Location     Location    `json:"location" binding:"required"`
	Interfaces   []Interface `json:"interfaces" binding:"required,min=1,dive"`
	SecureBoot   bool        `json:"secure_boot"`
	Virtualized  bool        `json:"virtualized"`
	SerialNumber string      `json:"serial_number" binding:"required"`
}

type UpdateNodeRequest struct {
	FarmID       uint64      `json:"farm_id" binding:"required,min=1"`
	Resources    Resources   `json:"resources" binding:"required,min=1"`
	Location     Location    `json:"location" binding:"required"`
	Interfaces   []Interface `json:"interfaces" binding:"required,dive"`
	SecureBoot   bool        `json:"secure_boot" binding:"required"`
	Virtualized  bool        `json:"virtualized" binding:"required"`
	SerialNumber string      `json:"serial_number" binding:"required"`
}

type UpdateAccountRequest struct {
	Relays    pq.StringArray `json:"relays"`
	RMBEncKey string         `json:"rmb_enc_key"`
}

type AccountCreationRequest struct {
	Timestamp int64  `json:"timestamp"`
	PublicKey string `json:"public_key"`
	// the registrar expect a signature of a message with format `timestampStr:publicKeyBase64`
	// - signature format: base64(ed25519_or_sr22519_signature)
	Signature string   `json:"signature"`
	Relays    []string `json:"relays"`
	RMBEncKey string   `json:"rmb_enc_key"`
}

type UptimeReportRequest struct {
	Uptime    time.Duration `json:"uptime"`
	Timestamp time.Time     `json:"timestamp"`
}

type Account struct {
	TwinID    uint64         `json:"twin_id"`
	Relays    pq.StringArray `json:"relays"`      // Optional list of relay domains
	RMBEncKey string         `json:"rmb_enc_key"` // Optional base64 encoded public key for rmb communication
	// The public key (ED25519 for nodes, ED25519 or SR25519 for farmers) in the more standard base64 since we are moving from substrate echo system?
	// (still SS58 can be used or plain base58 ,TBD)
	PublicKey string `json:"public_key"`
}

type Farm struct {
	FarmID    uint64 `json:"farm_id"`
	FarmName  string `json:"farm_name"`
	TwinID    uint64 `json:"twin_id"` // Farmer account reference
	Dedicated bool   `json:"dedicated"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Node struct {
	NodeID uint64 `json:"node_id"`
	FarmID uint64 `json:"farm_id"`
	TwinID uint64 `json:"twin_id"` // Node account reference

	Location Location `json:"location"`

	Resources    Resources   `json:"resources"`
	Interfaces   []Interface `json:"interface"`
	SecureBoot   bool
	Virtualized  bool
	SerialNumber string

	UptimeReports []UptimeReport `json:"uptime"`

	CreatedAt time.Time
	UpdatedAt time.Time
	Approved  bool
}

type UptimeReport struct {
	ID         uint64        `gorm:"primaryKey;autoIncrement"`
	NodeID     uint64        `gorm:"index"`
	Duration   time.Duration // Uptime duration for this period
	Timestamp  time.Time     `gorm:"index"`
	WasRestart bool          // True if this report followed a restart
	CreatedAt  time.Time
}

type ZosVersion struct {
	Key     string `gorm:"primaryKey;size:50"`
	Version string `gorm:"not null"`
}

type Interface struct {
	Name string `json:"name"`
	Mac  string `json:"mac"`
	IPs  string `json:"ips"`
}

type Resources struct {
	HRU uint64 `json:"hru"`
	SRU uint64 `json:"sru"`
	CRU uint64 `json:"cru"`
	MRU uint64 `json:"mru"`
}

type Location struct {
	Country   string `json:"country"`
	City      string `json:"city"`
	Longitude string `json:"longitude"`
	Latitude  string `json:"latitude"`
}
