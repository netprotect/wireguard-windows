module golang.zx2c4.com/wireguard/windows

require (
	github.com/lxn/walk v0.0.0-20190829150329-026611184a72
	github.com/lxn/win v0.0.0-20190716185335-d1d36f0e4f48

	golang.org/x/crypto v0.0.0-20190829043050-9756ffdc2472
	golang.org/x/net v0.0.0-20190827160401-ba9fcec4b297
	golang.org/x/sys v0.0.0-20190826190057-c7b8b68b1456
	golang.org/x/text v0.3.2
	golang.zx2c4.com/wireguard v0.0.20190806-0.20190829182215-fa7763c268d2
)

replace (
	github.com/lxn/walk => golang.zx2c4.com/wireguard/windows v0.0.0-20190829183229-2ce6866d7cc4
	github.com/lxn/win => golang.zx2c4.com/wireguard/windows v0.0.0-20190716185335-d1d36f0e4f48
)

go 1.13
