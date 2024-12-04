package xutil

func ChainIdGetTokenType(chainId string) string {
	switch chainId {
	case "11155111", "1", "97", "56", "88488848":
		return "ERC20"
	case "2494104990", "728126428", "2492104990":
		return "TRC20"
	case "555999555":
		return "SOL"
	default:
		return "NoType"
	}
}
