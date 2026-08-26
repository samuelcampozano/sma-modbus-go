package smamodbus

// Standard SMA Modbus Register addresses as defined in SMA Technical Information (EDMx-Modbus-TI).
const (
	// RegDeviceClass identifies the device category (U32, 2 regs). 8001 = Solar Inverter, 8128 = Data Manager.
	RegDeviceClass uint16 = 30051

	// RegSerialNumber represents the device serial number (U32, 2 regs).
	RegSerialNumber uint16 = 30057

	// RegOperatingStatus indicates the current operational condition (U32, 2 regs).
	// 35 = Fault, 303 = Off, 307 = OK, 455 = Warning.
	RegOperatingStatus uint16 = 30201

	// RegConnectedPower represents nominal connected DC/AC power in Watts (U32, 2 regs).
	RegConnectedPower uint16 = 30233

	// RegTotalEnergyFedIn represents lifetime cumulative feed-in energy in Watt-hours (U64, 4 regs).
	RegTotalEnergyFedIn uint16 = 30513

	// RegDailyYield represents today's energy yield in Watt-hours (U64, 4 regs).
	RegDailyYield uint16 = 30517

	// RegActivePowerTotal represents current instantaneous active PV power in Watts (S32, 2 regs).
	RegActivePowerTotal uint16 = 30775

	// RegActivePowerL1 represents current active power on phase L1 in Watts (S32, 2 regs).
	RegActivePowerL1 uint16 = 30777

	// RegActivePowerL2 represents current active power on phase L2 in Watts (S32, 2 regs).
	RegActivePowerL2 uint16 = 30779

	// RegActivePowerL3 represents current active power on phase L3 in Watts (S32, 2 regs).
	RegActivePowerL3 uint16 = 30781

	// RegGridFrequency represents AC grid frequency in Hz * 100 (U32, 2 regs).
	RegGridFrequency uint16 = 30803

	// RegReactivePower represents current reactive power in VAr (S32, 2 regs).
	RegReactivePower uint16 = 30805

	// RegInternalTemperature represents device internal temperature in °C * 10 (S32, 2 regs).
	RegInternalTemperature uint16 = 30953
)
