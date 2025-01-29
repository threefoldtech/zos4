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
	Timestamp int64  `json:"timestamp" binding:"required"`
	PublicKey string `json:"public_key" binding:"required"` // base64 encoded
	// the registrar expect a signature of a message with format `timestampStr:publicKeyBase64`
	// - signature format: base64(ed25519_or_sr22519_signature)
	Signature string   `json:"signature" binding:"required"`
	Relays    []string `json:"relays,omitempty"`
	RMBEncKey string   `json:"rmb_enc_key,omitempty"`
}
type UptimeReportRequest struct {
	Uptime    time.Duration `json:"uptime" binding:"required"`
	Timestamp time.Time     `json:"timestamp" binding:"required"`
}
type Account struct {
	TwinID    uint64         `gorm:"primaryKey;autoIncrement"`
	Relays    pq.StringArray `gorm:"type:text[];default:'{}'" json:"relays"` // Optional list of relay domains
	RMBEncKey string         `gorm:"type:text" json:"rmb_enc_key"`           // Optional base64 encoded public key for rmb communication
	CreatedAt time.Time
	UpdatedAt time.Time
	// The public key (ED25519 for nodes, ED25519 or SR25519 for farmers) in the more standard base64 since we are moving from substrate echo system?
	// (still SS58 can be used or plain base58 ,TBD)
	PublicKey string `gorm:"type:text;not null;unique"`
	// Relations | likely we need to use OnDelete:RESTRICT (Prevent Twin deletion if farms exist)
	Farms []Farm `gorm:"foreignKey:TwinID;references:TwinID;constraint:OnDelete:RESTRICT"`
}

type Farm struct {
	FarmID    uint64 `gorm:"primaryKey;autoIncrement" json:"farm_id"`
	FarmName  string `gorm:"size:40;not null;unique;check:farm_name <> ''" json:"farm_name"`
	TwinID    uint64 `json:"twin_id" gorm:"not null;check:twin_id > 0"` // Farmer account reference
	Dedicated bool   `json:"dedicated"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Nodes []Node `gorm:"foreignKey:FarmID;references:FarmID;constraint:OnDelete:RESTRICT" json:"nodes"`
}

type Node struct {
	NodeID uint64 `json:"node_id" gorm:"primaryKey;autoIncrement"`
	// Constraints set to prevents unintended account deletion if linked Farms/nodes exist.
	FarmID uint64 `json:"farm_id" gorm:"not null;check:farm_id> 0;foreignKey:FarmID;references:FarmID;constraint:OnDelete:RESTRICT"`
	TwinID uint64 `json:"twin_id" gorm:"not null;unique;check:twin_id > 0;foreignKey:TwinID;references:TwinID;constraint:OnDelete:RESTRICT"` // Node account reference

	Location Location `json:"location" gorm:"not null;type:json"`

	// PublicConfig PublicConfig `json:"public_config" gorm:"type:json"`
	Resources    Resources   `json:"resources" gorm:"not null;type:json"`
	Interfaces   []Interface `json:"interface" gorm:"not null;type:json"`
	SecureBoot   bool
	Virtualized  bool
	SerialNumber string

	UptimeReports []UptimeReport `json:"uptime" gorm:"foreignKey:NodeID;references:NodeID;constraint:OnDelete:CASCADE"`

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
	Country   string `json:"country" gorm:"not null"`
	City      string `json:"city" gorm:"not null"`
	Longitude string `json:"longitude" gorm:"not null"`
	Latitude  string `json:"latitude" gorm:"not null"`
}
