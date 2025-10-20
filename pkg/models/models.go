package models

type Kismet struct {
	KismetVersion string `gorm:"column:kismet_version"`
	DBVersion     int    `gorm:"column:db_version"`
	DBModule      string `gorm:"column:db_module"`
}

func (Kismet) TableName() string { return "KISMET" }

type Device struct {
	FirstTime       int     `gorm:"column:first_time"`
	LastTime        int     `gorm:"column:last_time"`
	DevKey          string  `gorm:"column:devkey"`
	PhyName         string  `gorm:"column:phyname;index:idx_phy_devmac,unique"`
	DevMac          string  `gorm:"column:devmac;index:idx_phy_devmac,unique"`
	StrongestSignal int     `gorm:"column:strongest_signal"`
	MinLat          float64 `gorm:"column:min_lat"`
	MinLon          float64 `gorm:"column:min_lon"`
	MaxLat          float64 `gorm:"column:max_lat"`
	MaxLon          float64 `gorm:"column:max_lon"`
	AvgLat          float64 `gorm:"column:avg_lat"`
	AvgLon          float64 `gorm:"column:avg_lon"`
	BytesData       int     `gorm:"column:bytes_data"`
	Type            string  `gorm:"column:type"`
	DeviceBlob      []byte  `gorm:"column:device"`
}

func (Device) TableName() string { return "devices" }

type Packet struct {
	TsSec      int     `gorm:"column:ts_sec"`
	TsUsec     int     `gorm:"column:ts_usec"`
	PhyName    string  `gorm:"column:phyname"`
	SourceMac  string  `gorm:"column:sourcemac"`
	DestMac    string  `gorm:"column:destmac"`
	TransMac   string  `gorm:"column:transmac"`
	Frequency  float64 `gorm:"column:frequency"`
	DevKey     string  `gorm:"column:devkey"`
	Lat        float64 `gorm:"column:lat"`
	Lon        float64 `gorm:"column:lon"`
	Alt        float64 `gorm:"column:alt"`
	Speed      float64 `gorm:"column:speed"`
	Heading    float64 `gorm:"column:heading"`
	PacketLen  int     `gorm:"column:packet_len"`
	Signal     int     `gorm:"column:signal"`
	DataSource string  `gorm:"column:datasource"`
	Dlt        int     `gorm:"column:dlt"`
	PacketBlob []byte  `gorm:"column:packet"`
	Error      int     `gorm:"column:error"`
	Tags       string  `gorm:"column:tags"`
	DataRate   float64 `gorm:"column:datarate"`
	Hash       int     `gorm:"column:hash"`
	PacketID   int     `gorm:"column:packetid"`
}

func (Packet) TableName() string { return "packets" }

type Data struct {
	TsSec      int     `gorm:"column:ts_sec"`
	TsUsec     int     `gorm:"column:ts_usec"`
	PhyName    string  `gorm:"column:phyname"`
	DevMac     string  `gorm:"column:devmac"`
	Lat        float64 `gorm:"column:lat"`
	Lon        float64 `gorm:"column:lon"`
	Alt        float64 `gorm:"column:alt"`
	Speed      float64 `gorm:"column:speed"`
	Heading    float64 `gorm:"column:heading"`
	DataSource string  `gorm:"column:datasource"`
	Type       string  `gorm:"column:type"`
	JSONBlob   []byte  `gorm:"column:json"`
}

func (Data) TableName() string { return "data" }

type DataSource struct {
	UUID       string `gorm:"column:uuid;uniqueIndex"`
	TypeString string `gorm:"column:typestring"`
	Definition string `gorm:"column:definition"`
	Name       string `gorm:"column:name"`
	Interface  string `gorm:"column:interface"`
	JSONBlob   []byte `gorm:"column:json"`
}

func (DataSource) TableName() string { return "datasources" }

type Alert struct {
	TsSec    int     `gorm:"column:ts_sec"`
	TsUsec   int     `gorm:"column:ts_usec"`
	PhyName  string  `gorm:"column:phyname"`
	DevMac   string  `gorm:"column:devmac"`
	Lat      float64 `gorm:"column:lat"`
	Lon      float64 `gorm:"column:lon"`
	Header   string  `gorm:"column:header"`
	JSONBlob []byte  `gorm:"column:json"`
}

func (Alert) TableName() string { return "alerts" }

type Message struct {
	TsSec   int     `gorm:"column:ts_sec"`
	Lat     float64 `gorm:"column:lat"`
	Lon     float64 `gorm:"column:lon"`
	MsgType string  `gorm:"column:msgtype"`
	Message string  `gorm:"column:message"`
}

func (Message) TableName() string { return "messages" }

type Snapshot struct {
	TsSec    int     `gorm:"column:ts_sec"`
	TsUsec   int     `gorm:"column:ts_usec"`
	Lat      float64 `gorm:"column:lat"`
	Lon      float64 `gorm:"column:lon"`
	SnapType string  `gorm:"column:snaptype"`
	JSONBlob []byte  `gorm:"column:json"`
}

func (Snapshot) TableName() string { return "snapshots" }
