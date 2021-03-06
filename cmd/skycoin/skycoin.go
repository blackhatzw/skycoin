package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/op/go-logging"
)

import (
	//"github.com/skycoin/skycoin/src/cli"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/daemon"
	"github.com/skycoin/skycoin/src/gui"
	"github.com/skycoin/skycoin/src/util"
	"github.com/skycoin/skycoin/src/wallet"
)

// Command line interface arguments

type Args interface {
	register()
	postProcess()
	getConfig() *Config
}

type Config struct {
	DisableGUI bool
	// Disable DHT peer discovery
	DisableDHT bool
	// Disable peer exchange
	DisablePEX bool
	// Don't make any outgoing connections
	DisableOutgoingConnections bool
	// Don't allowing incoming connections
	DisableIncomingConnections bool
	// Disables networking altogether
	DisableNetworking bool
	// Only run on localhost and only connect to others on localhost
	LocalhostOnly bool
	// Which address to serve on. Leave blank to automatically assign to a
	// public interface
	Address string
	// DHT uses this port for UDP; gnet uses this for TCP incoming and outgoing
	Port int
	// How often to make outgoing connections
	OutgoingConnectionsRate time.Duration
	// Wallet Address Version
	AddressVersion string
	// Remote web interface
	WebInterface      bool
	WebInterfacePort  int
	WebInterfaceAddr  string
	WebInterfaceCert  string
	WebInterfaceKey   string
	WebInterfaceHTTPS bool
	// Data directory holds app data -- defaults to ~/.skycoin
	DataDirectory string
	// GUI directory contains assets for the html gui
	GUIDirectory string
	// Logging
	LogLevel logging.Level
	ColorLog bool
	// This is the value registered with flag, it is converted to LogLevel
	// after parsing
	logLevel string

	// Wallets
	WalletDirectory string
	BlockchainFile  string
	BlockSigsFile   string
	// Is allowed to make outgoing transactions
	CanSpend bool

	// Centralized network configuration
	BlockchainPubkey string
	RunMaster        bool
	MasterKeys       string
	GenesisSignature string
	GenesisTimestamp uint64

	/* Developer options */

	// Enable cpu profiling
	ProfileCPU bool
	// Where the file is written to
	ProfileCPUFile string
	// HTTP profiling interface (see http://golang.org/pkg/net/http/pprof/)
	HTTPProf bool
	// Will force it to connect to this ip:port, instead of waiting for it
	// to show up as a peer
	ConnectTo string
}

func (self *Config) register() {
	log.Panic("Config.register must be overridden")
}

func (self *Config) postProcess() {
	self.DataDirectory = util.InitDataDir(self.DataDirectory)
	if self.WebInterfaceCert == "" {
		self.WebInterfaceCert = filepath.Join(self.DataDirectory, "cert.pem")
	}
	if self.WebInterfaceKey == "" {
		self.WebInterfaceKey = filepath.Join(self.DataDirectory, "key.pem")
	}
	if self.MasterKeys == "" {
		self.MasterKeys = filepath.Join(self.DataDirectory, "master.keys")
	}
	if self.BlockchainFile == "" {
		self.BlockchainFile = filepath.Join(self.DataDirectory, "blockchain.bin")
	}
	if self.BlockSigsFile == "" {
		self.BlockSigsFile = filepath.Join(self.DataDirectory, "blockchain.sigs")
	}
	if self.WalletDirectory == "" {
		self.WalletDirectory = filepath.Join(self.DataDirectory, "wallets/")
	}
	ll, err := logging.LogLevel(self.logLevel)
	if err != nil {
		log.Panic("Invalid -log-level %s: %v\n", self.logLevel, err)
	}
	self.LogLevel = ll
}

func (self *Config) getConfig() *Config {
	return self
}

// Parses arguments defined in a struct that satisfies Config interface
func ParseArgs(args Args) *Config {
	args.register()
	flag.Parse()
	args.postProcess()
	return args.getConfig()
}

/*
 Dev Args
*/

type DevConfig struct {
	Config
}

var DevArgs = DevConfig{Config{
	DisableGUI: true,
	// Disable DHT peer discovery
	DisableDHT: false,
	// Disable peer exchange
	DisablePEX: false,
	// Don't make any outgoing connections
	DisableOutgoingConnections: false,
	// Don't allowing incoming connections
	DisableIncomingConnections: false,
	// Disables networking altogether
	DisableNetworking: false,
	// Only run on localhost and only connect to others on localhost
	LocalhostOnly: false,
	// Which address to serve on. Leave blank to automatically assign to a
	// public interface
	Address: "",
	// DHT uses this port for UDP; gnet uses this for TCP incoming and outgoing
	Port: 5798,
	// How often to make outgoing connections, in seconds
	OutgoingConnectionsRate: time.Second * 5,
	// Wallet Address Version
	AddressVersion: "test",
	// Remote web interface
	WebInterface:      false,
	WebInterfacePort:  6402,
	WebInterfaceAddr:  "127.0.0.1",
	WebInterfaceCert:  "",
	WebInterfaceKey:   "",
	WebInterfaceHTTPS: false,
	// Data directory holds app data -- defaults to ~/.skycoin
	DataDirectory: "",
	// Data directory holds app data -- defaults to ~/.skycoin
	GUIDirectory: "./src/gui/static/",
	// Logging
	LogLevel: logging.DEBUG,
	ColorLog: true,
	logLevel: "DEBUG",

	// Wallets
	WalletDirectory: "",
	BlockchainFile:  "",
	BlockSigsFile:   "",
	CanSpend:        true,

	// Centralized network configuration
	RunMaster:        true,
	BlockchainPubkey: cipher.MustPubKeyFromHex("02b0333bd8f1910663b8b1f60fb2e154b70436a2c19efb79cdbdf09bf9bb2056dc"),
	BlockchainSeckey: cipher.MustSecKeyFromHex("02b0333bd8f1910663b8b1f60fb2e154b70436a2c19efb79cdbdf09bf9bb2056dc"),

	MasterKeys:       "",
	GenesisTimestamp: 1394689119,
	GenesisSignature: "173e1cdf628e78ae4946af4415f070e2aad5a1f4273b77971f8d42a6eb7ff3af68d0d7a3360460e96123f93decf43c28abbc02a65ffb243e525131ba357f21d800",

	/* Developer options */

	// Enable cpu profiling
	ProfileCPU: false,
	// Where the file is written to
	ProfileCPUFile: "skycoin.prof",
	// HTTP profiling interface (see http://golang.org/pkg/net/http/pprof/)
	HTTPProf: false,
	// Will force it to connect to this ip:port, instead of waiting for it
	// to show up as a peer
	ConnectTo: "",
}}

func (self *DevConfig) register() {
	flag.BoolVar(&self.DisableDHT, "disable-dht", self.DisableDHT,
		"disable DHT peer discovery")
	flag.BoolVar(&self.DisablePEX, "disable-pex", self.DisablePEX,
		"disable PEX peer discovery")
	flag.BoolVar(&self.DisableOutgoingConnections, "disable-outgoing",
		self.DisableOutgoingConnections, "Don't make outgoing connections")
	flag.BoolVar(&self.DisableIncomingConnections, "disable-incoming",
		self.DisableIncomingConnections, "Don't make incoming connections")
	flag.BoolVar(&self.DisableNetworking, "disable-networking",
		self.DisableNetworking, "Disable all network activity")
	flag.StringVar(&self.Address, "address", self.Address,
		"IP Address to run application on. Leave empty to default to a public interface")
	flag.IntVar(&self.Port, "port", self.Port, "Port to run application on")
	flag.BoolVar(&self.DisableGUI, "disable-gui", self.DisableGUI,
		"disable the gui")
	flag.BoolVar(&self.WebInterface, "web-interface", self.WebInterface,
		"enable the web interface")
	flag.IntVar(&self.WebInterfacePort, "web-interface-port",
		self.WebInterfacePort, "port to serve web interface on")
	flag.StringVar(&self.WebInterfaceAddr, "web-interface-addr",
		self.WebInterfaceAddr, "addr to serve web interface on")
	flag.StringVar(&self.WebInterfaceCert, "web-interface-cert",
		self.WebInterfaceCert, "cert.pem file for web interface HTTPS. "+
			"If not provided, will use cert.pem in -data-directory")
	flag.StringVar(&self.WebInterfaceKey, "web-interface-key",
		self.WebInterfaceKey, "key.pem file for web interface HTTPS. "+
			"If not provided, will use key.pem in -data-directory")
	flag.BoolVar(&self.WebInterfaceHTTPS, "web-interface-https",
		self.WebInterfaceHTTPS, "enable HTTPS for web interface")
	flag.StringVar(&self.DataDirectory, "data-dir", self.DataDirectory,
		"directory to store app data (defaults to ~/.skycoin)")
	flag.StringVar(&self.ConnectTo, "connect-to", self.ConnectTo,
		"connect to this ip only")
	flag.BoolVar(&self.ProfileCPU, "profile-cpu", self.ProfileCPU,
		"enable cpu profiling")
	flag.StringVar(&self.ProfileCPUFile, "profile-cpu-file",
		self.ProfileCPUFile, "where to write the cpu profile file")
	flag.BoolVar(&self.HTTPProf, "http-prof", self.HTTPProf,
		"Run the http profiling interface")
	flag.StringVar(&self.logLevel, "log-level", self.logLevel,
		"Choices are: debug, info, notice, warning, error, critical")
	flag.BoolVar(&self.ColorLog, "color-log", self.ColorLog,
		"Add terminal colors to log output")
	flag.StringVar(&self.GUIDirectory, "gui-dir", self.GUIDirectory,
		"static content directory for the html gui")
	flag.BoolVar(&self.RunMaster, "master", self.RunMaster,
		"run the daemon as blockchain master server")
	flag.StringVar(&self.MasterKeys, "master-keys", self.MasterKeys,
		"file to load master keys and address from")
	flag.StringVar(&self.BlockchainPubkey, "master-public-key", self.BlockchainPubkey,
		"public key of the master chain")
	flag.StringVar(&self.GenesisSignature, "genesis-signature", self.GenesisSignature,
		"genesis block signature")
	flag.Uint64Var(&self.GenesisTimestamp, "genesis-timestamp", self.GenesisTimestamp,
		"genesis block timestamp")
	flag.StringVar(&self.WalletDirectory, "wallet-dir", self.WalletDirectory,
		"location of the wallet files. Defaults to ~/.skycoin/wallet/")
	flag.StringVar(&self.BlockchainFile, "blockchain-file", self.BlockchainFile,
		"location of the blockchain file. Default to ~/.skycoin/blockchain.bin")
	flag.StringVar(&self.BlockSigsFile, "blocksigs-file", self.BlockSigsFile,
		"location of the block signatures file. Default to ~/.skycoin/blockchain.sigs")
	flag.BoolVar(&self.CanSpend, "can-spend", self.CanSpend,
		"is allowed to make outgoing transactions")
	flag.DurationVar(&self.OutgoingConnectionsRate, "connection-rate",
		self.OutgoingConnectionsRate, "How often to make an outgoing connection")
	flag.BoolVar(&self.LocalhostOnly, "localhost-only", self.LocalhostOnly,
		"Run on localhost and only connect to localhost peers")
	flag.StringVar(&self.AddressVersion, "address-version", self.AddressVersion,
		"Wallet address version. Options are 'test' and 'main'")
}

/*
End Dev Args
*/
var (
	logger     = logging.MustGetLogger("skycoin.main")
	logFormat  = "[%{module}:%{level}] %{message}"
	logModules = []string{
		"skycoin.main",
		"skycoin.daemon",
		"skycoin.coin",
		"skycoin.gui",
		"skycoin.util",
		"skycoin.visor",
		"skycoin.wallet",
		"gnet",
		"pex",
	}
)

func printProgramStatus() {
	fn := "goroutine.prof"
	logger.Debug("Writing goroutine profile to %s", fn)
	p := pprof.Lookup("goroutine")
	f, err := os.Create(fn)
	defer f.Close()
	if err != nil {
		logger.Error("%v", err)
		return
	}
	err = p.WriteTo(f, 2)
	if err != nil {
		logger.Error("%v", err)
		return
	}
}

func catchInterrupt(quit chan<- int) {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	<-sigchan
	signal.Stop(sigchan)
	quit <- 1
}

// Catches SIGUSR1 and prints internal program state
func catchDebug() {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGUSR1)
	for {
		select {
		case <-sigchan:
			printProgramStatus()
		}
	}
}

// func initSettings() {
//     sb.InitSettings()
//     sb.Settings.Load()
//     we resave the settings, in case they were not found and had to be generated
//     sb.Settings.Save()
// }

func initLogging(level logging.Level, color bool) {
	format := logging.MustStringFormatter(logFormat)
	logging.SetFormatter(format)
	for _, s := range logModules {
		logging.SetLevel(level, s)
	}
	stdout := logging.NewLogBackend(os.Stdout, "", 0)
	stdout.Color = color
	logging.SetBackend(stdout)
}

func initProfiling(httpProf, profileCPU bool, profileCPUFile string) {
	if profileCPU {
		f, err := os.Create(profileCPUFile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	if httpProf {
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}
}

func configureDaemon(c *Config) daemon.Config {
	//cipher.SetAddressVersion(c.AddressVersion)
	dc := daemon.NewConfig()
	dc.Peers.DataDirectory = c.DataDirectory
	dc.DHT.Disabled = c.DisableDHT
	dc.Peers.Disabled = c.DisablePEX
	dc.Daemon.DisableOutgoingConnections = c.DisableOutgoingConnections
	dc.Daemon.DisableIncomingConnections = c.DisableIncomingConnections
	dc.Daemon.DisableNetworking = c.DisableNetworking
	dc.Daemon.Port = c.Port
	dc.Daemon.Address = c.Address
	dc.Daemon.LocalhostOnly = c.LocalhostOnly
	if c.OutgoingConnectionsRate == 0 {
		c.OutgoingConnectionsRate = time.Millisecond
	}
	dc.Daemon.OutgoingRate = c.OutgoingConnectionsRate

	dc.Visor.Config.WalletDirectory = c.WalletDirectory
	dc.Visor.Config.BlockchainFile = c.BlockchainFile
	dc.Visor.Config.BlockSigsFile = c.BlockSigsFile

	dc.Visor.Config.IsMaster = c.RunMaster

	//generate new private/public key
	pub, sec := cipher.GenerateDeterministicKeyPair([]byte("genesis"))

	dc.Visor.Config.BlockchainPubkey = pub
	dc.Visor.Config.BlockchainSeckey = sec
	dc.Visor.Config.GenesisAddress = cipher.AddressFromPubKey(c.BlockchainPubkey)

	dc.Visor.Config.GenesisSignature = cipher.Sig{}
	dc.Visor.Config.GenesisTimestamp = c.GenesisTimestamp
	dc.Visor.Config.WalletConstructor = wallet.NewDeterministicWallet

	/*
		dc.Visor.MasterKeysFile = c.MasterKeys
		if c.RunMaster {
			// Will panic if fails
			//dc.Visor.LoadMasterKeys()
		} else {
			//w := wallet.ReadableWalletEntryFromPubkey(c.BlockchainPubkey)
			//dc.Visor.Config.MasterKeys = wallet.WalletEntryFromReadable(&w)
		}
	*/

	return dc
}

func Run(args Args) {
	c := ParseArgs(args)
	initProfiling(c.HTTPProf, c.ProfileCPU, c.ProfileCPUFile)
	initLogging(c.LogLevel, c.ColorLog)

	// If the user Ctrl-C's, shutdown properly
	quit := make(chan int)
	go catchInterrupt(quit)
	// Watch for SIGUSR1
	go catchDebug()

	err := os.MkdirAll(c.WalletDirectory, os.FileMode(0700))
	if err != nil {
		logger.Critical("Failed to create wallet directory: %v", err)
	}

	dconf := configureDaemon(c)
	d := daemon.NewDaemon(dconf)

	stopDaemon := make(chan int)
	go d.Start(stopDaemon)

	// Debug only - forces connection on start.  Violates thread safety.
	if c.ConnectTo != "" {
		_, err := d.Pool.Pool.Connect(c.ConnectTo)
		if err != nil {
			log.Panic(err)
		}
	}

	if !c.DisableGUI {
		go gui.LaunchGUI(d)
	}

	host := fmt.Sprintf("%s:%d", c.WebInterfaceAddr, c.WebInterfacePort)

	if c.WebInterface {
		if c.WebInterfaceHTTPS {
			// Verify cert/key parameters, and if neither exist, create them
			errs := gui.CreateCertIfNotExists(host, c.WebInterfaceCert,
				c.WebInterfaceKey)
			if len(errs) != 0 {
				for _, err := range errs {
					logger.Error(err.Error())
				}
			} else {
				go gui.LaunchWebInterfaceHTTPS(host, c.GUIDirectory, d,
					c.WebInterfaceCert, c.WebInterfaceKey)
			}
		} else {
			go gui.LaunchWebInterface(host, c.GUIDirectory, d)
		}
	}

	<-quit
	stopDaemon <- 1

	logger.Info("Shutting down")
	d.Shutdown()
	logger.Info("Goodbye")
}

func main() {
	/*
		skycoin.Run(&cli.DaemonArgs)
	*/

	/*
	   skycoin.Run(&cli.ClientArgs)
	   stop := make(chan int)
	   <-stop
	*/

	//skycoin.Run(&cli.DevArgs)
	Run(&DevArgs)
}
