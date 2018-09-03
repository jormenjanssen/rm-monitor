package main

// BroadbandConnType connection type
type BroadbandConnType int

const (
	// ConnTypeNoNetwork means not connected to any network or unkown
	ConnTypeNoNetwork BroadbandConnType = 0
	// ConnType2G means GRPS or Edge connection
	ConnType2G BroadbandConnType = 1
	// ConnType3G means HSDPA or HSUPA, etc
	ConnType3G BroadbandConnType = 2
	// ConnType4G means lte or other
	ConnType4G BroadbandConnType = 3
)
