package wallet

import (
	"encoding/hex"
	"fmt"
	"log"
	"path/filepath"

	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/cipher/secp256k1-go"
)

// Simplest wallet implementation
type SimpleWallet struct {
	ID       WalletID
	Name     string
	Filename string
	Entries  WalletEntries
}

func NewEmptySimpleWallet() Wallet {
	idHash := cipher.SumSHA256(secp256k1.RandByte(256))
	id := WalletID(hex.EncodeToString(idHash[:16]))
	return &SimpleWallet{
		Filename: NewWalletFilename(id),
		Entries:  WalletEntries{},
		ID:       id,
	}
}

func NewSimpleWallet() Wallet {
	w := NewEmptySimpleWallet().(*SimpleWallet)
	e := NewWalletEntry()
	w.Entries = WalletEntries{e.Address: e}
	return w
}

func LoadSimpleWallet(dir, filename string) (Wallet, error) {
	w := &SimpleWallet{
		Filename: filename,
		Entries:  make(WalletEntries),
	}
	return w, w.Load(dir)
}

func NewSimpleWalletFromReadable(r *ReadableWallet) Wallet {
	return &SimpleWallet{
		Name:     r.Name,
		Filename: r.Filename,
		Entries:  r.Entries.ToWalletEntries(),
		ID:       WalletID(r.Extra["id"].(string)),
	}
}

func (self *SimpleWallet) GetType() WalletType {
	return SimpleWalletType
}

func (self *SimpleWallet) GetFilename() string {
	return self.Filename
}

func (self *SimpleWallet) SetFilename(fn string) {
	self.Filename = fn
}

func (self *SimpleWallet) GetID() WalletID {
	return self.ID
}

func (self *SimpleWallet) GetName() string {
	return self.Name
}

func (self *SimpleWallet) SetName(name string) {
	self.Name = name
}

func (self *SimpleWallet) NumEntries() int {
	return len(self.Entries)
}

func (self *SimpleWallet) GetEntries() WalletEntries {
	return self.Entries
}

func (self *SimpleWallet) GetAddressSet() AddressSet {
	m := make(AddressSet, len(self.Entries))
	for a, _ := range self.Entries {
		m[a] = byte(1)
	}
	return m
}

// Creates a WalletEntry
func (self *SimpleWallet) CreateEntry() WalletEntry {
	e := NewWalletEntry()
	if err := self.AddEntry(e); err != nil {
		log.Panic("Somehow, we managed to create a bad entry: %v", err)
	}
	return e
}

// Returns all cipher.Addresses in this SimpleWallet
func (self *SimpleWallet) GetAddresses() []cipher.Address {
	addrs := make([]cipher.Address, len(self.Entries))
	i := 0
	for a, _ := range self.Entries {
		addrs[i] = a
		i++
	}
	return addrs
}

// Returns the WalletEntry for a cipher.Address
func (self *SimpleWallet) GetEntry(a cipher.Address) (WalletEntry, bool) {
	we, exists := self.Entries[a]
	return we, exists
}

// Adds a WalletEntry to the wallet. Returns an error if the cipher.Address is
// already in the wallet
func (self *SimpleWallet) AddEntry(e WalletEntry) error {
	if err := e.Verify(); err != nil {
		return err
	}
	_, exists := self.Entries[e.Address]
	if exists {
		return fmt.Errorf("SimpleWallet entry already exists for address %s",
			e.Address.String())
	} else {
		self.Entries[e.Address] = e
		return nil
	}
}

// Saves to filename
func (self *SimpleWallet) Save(dir string) error {
	r := NewReadableWallet(self)
	return r.Save(filepath.Join(dir, self.Filename))
}

// Loads from filename
func (self *SimpleWallet) Load(dir string) error {
	fn := self.Filename
	r, err := LoadReadableWallet(filepath.Join(dir, fn))
	if err != nil {
		return err
	}
	*self = *(NewSimpleWalletFromReadable(r)).(*SimpleWallet)
	self.Filename = fn
	return nil
}

func (self *SimpleWallet) GetExtraSerializerData() map[string]interface{} {
	m := make(map[string]interface{}, 1)
	m["id"] = string(self.ID)
	return m
}
