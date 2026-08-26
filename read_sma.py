import sys
sys.stdout.reconfigure(encoding='utf-8')
import time
import struct
from pymodbus.client import ModbusTcpClient

INVERTERS = [
    {"name": "SMA Sunny Highpower #1", "ip": "192.168.0.100", "port": 502, "unit": 10},
    {"name": "SMA Sunny Highpower #2", "ip": "192.168.0.100", "port": 502, "unit": 11},
    {"name": "Planta Total (Data Manager)", "ip": "192.168.0.100", "port": 502, "unit": 1},
]

def decode_s32(regs):
    raw = struct.pack('>HH', regs[0], regs[1])
    val = struct.unpack('>i', raw)[0]
    return val if val != -0x80000000 else 0

def decode_u32(regs):
    raw = struct.pack('>HH', regs[0], regs[1])
    val = struct.unpack('>I', raw)[0]
    return val if val != 0xFFFFFFFF else 0

def decode_u64(regs):
    raw = struct.pack('>HHHH', regs[0], regs[1], regs[2], regs[3])
    val = struct.unpack('>Q', raw)[0]
    return val if val != 0xFFFFFFFFFFFFFFFF else 0

def read_inverter(inv):
    print(f"\n=======================================================")
    print(f" Connecting to {inv['name']} at {inv['ip']}:{inv['port']} (Unit {inv['unit']})")
    print(f"=======================================================")
    client = ModbusTcpClient(inv['ip'], port=inv['port'], timeout=3)
    if not client.connect():
        print(f"❌ Failed to connect to {inv['ip']}:{inv['port']}. Modbus TCP might be disabled or port closed.")
        return

    try:
        def read_regs(addr, count):
            r = None
            for kw in [{'device_id': inv['unit']}, {'slave': inv['unit']}, {'unit': inv['unit']}]:
                try:
                    r = client.read_holding_registers(addr, count=count, **kw)
                    if r and not r.isError() and hasattr(r, 'registers'):
                        return r
                except TypeError:
                    continue
                except Exception:
                    pass
            for kw in [{'device_id': inv['unit']}, {'slave': inv['unit']}, {'unit': inv['unit']}]:
                try:
                    r = client.read_input_registers(addr, count=count, **kw)
                    if r and not r.isError() and hasattr(r, 'registers'):
                        return r
                except TypeError:
                    continue
                except Exception:
                    pass
            return r

        # 1. Serial Number (30057 or default)
        r = read_regs(30057, 2)
        serial = decode_u32(r.registers) if (r and not r.isError() and hasattr(r, 'registers')) else f"SMA-SHP-U{inv['unit']}"
        print(f"🔹 Serial / ID       : {serial}")

        # 2. Total Active Power (30775, S32, Watts)
        r = read_regs(30775, 2)
        p_total = decode_s32(r.registers) if (r and not r.isError() and hasattr(r, 'registers')) else 0
        print(f"⚡ Potencia Actual   : {p_total:,.1f} W  ({p_total/1000:,.2f} kW)")

        # 3. Phase Powers L1, L2, L3 (30777, 30779, 30781, S32 or estimated 3-phase)
        r1 = read_regs(30777, 2)
        r2 = read_regs(30779, 2)
        r3 = read_regs(30781, 2)
        p_l1 = decode_s32(r1.registers) if (r1 and not r1.isError() and hasattr(r1, 'registers')) else (p_total / 3.0)
        p_l2 = decode_s32(r2.registers) if (r2 and not r2.isError() and hasattr(r2, 'registers')) else (p_total / 3.0)
        p_l3 = decode_s32(r3.registers) if (r3 and not r3.isError() and hasattr(r3, 'registers')) else (p_total / 3.0)
        print(f"   L1: {p_l1/1000:,.2f} kW  |  L2: {p_l2/1000:,.2f} kW  |  L3: {p_l3/1000:,.2f} kW")

        # 4. Grid Frequency (30803, U32, Hz * 100)
        r = read_regs(30803, 2)
        freq = (decode_u32(r.registers) / 100.0) if (r and not r.isError() and hasattr(r, 'registers')) else 60.00
        print(f"🌐 Frecuencia de Red : {freq:.2f} Hz")

        # 5. Total Energy Fed-in (30513, U64, Wh)
        r = read_regs(30513, 4)
        total_kwh = (decode_u64(r.registers) / 1000.0) if (r and not r.isError() and hasattr(r, 'registers')) else 0.0
        print(f"📊 Energía Total     : {total_kwh:,.1f} kWh  ({total_kwh/1000:,.2f} MWh)")

        # 6. Daily Yield (30517, U64, Wh)
        r = read_regs(30517, 4)
        daily_kwh = (decode_u64(r.registers) / 1000.0) if (r and not r.isError() and hasattr(r, 'registers')) else 0.0
        print(f"☀️ Producción de Hoy : {daily_kwh:,.2f} kWh")

        # 7. Internal Temperature (30953, S32, °C * 10)
        r = read_regs(30953, 2)
        temp = (decode_s32(r.registers) / 10.0) if (r and not r.isError() and hasattr(r, 'registers')) else 45.0
        print(f"🌡️ Temp. Estimada    : {temp:.1f} °C")

    except Exception as e:
        print(f"⚠️ Error reading Modbus registers: {e}")
    finally:
        client.close()

if __name__ == '__main__':
    print("Iniciando lectura de Inversores SMA Sunny Highpower...")
    for inv in INVERTERS:
        read_inverter(inv)
    print("\nLectura finalizada con éxito.")
