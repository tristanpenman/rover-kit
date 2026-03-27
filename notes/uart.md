# UART

This document defines the binary protocol exchanged between STM32 sonar firmware and the host-side decoder.

## Goals

- Keep framing self-synchronizing on a byte stream (UART).
- Include a protocol version in every frame for backward/forward compatibility.
- Detect corruption with CRC and recover quickly without restarting the stream.

## Frame format

All multi-byte integers are little-endian.

Frame structure:

```
+-------------+--------------+--------------------------------------------------------------+
| Field       | Size (bytes) | Description                                                  |
+-------------+--------------+--------------------------------------------------------------+
| start_0     |            1 | Constant 0xAA                                                |
| start_1     |            1 | Constant 0x55                                                |
| version     |            1 | Protocol version. Current version: 0x01                      |
| payload_len |            1 | Number of payload bytes that follow                          |
| payload     |  payload_len | Version-specific payload                                     |
| crc16       |            2 | CRC-16/CCITT-FALSE over version || payload_len || payload    |
+-------------+--------------+--------------------------------------------------------------+
```

### CRC details

- Algorithm: CRC-16/CCITT-FALSE
- Polynomial: `0x1021`
- Initial value: `0xFFFF`
- No reflection
- Final XOR: `0x0000`
- On-wire byte order: little-endian (`crc_lo`, then `crc_hi`)

## Version `0x01` payload

Constraints for version `0x01`:

- Payload length must be exactly `6 + (2 * sensor_count)`.
- Sensor count should match the number of active sonar channels configured in firmware.

Payload structure:

```
+---------------+--------------+-------------------------------------------------------------+
| Field         | Size (bytes) | Description                                                 |
+---------------+--------------+-------------------------------------------------------------+
| timestamp_ms  |            4 | Milliseconds since STM32 boot; wraps at uint32              |
| distance_unit |            1 | 0x00 = millimeters, 0x01 = centimeters                      |
| sensor_count  |            1 | Number of distance readings in frame (N)                    |
| readings      |        2 * N | Unsigned uint16 distance values in distance_unit            |
+---------------+--------------+-------------------------------------------------------------+
```

## Units and timestamp semantics

- Firmware SHOULD emit millimeters (`distance_unit = 0x00`) for maximal precision.
- Host may convert to centimeters for display, but transport preserves the native unit.
- `timestamp_ms` is monotonic relative to firmware boot time and not wall-clock time.
- Host code must handle timestamp wrap-around (every ~49.7 days).

## Error handling and resynchronization behavior

On the host side, decoder behavior is:

1. **Wait for start bytes**: bytes before `0xAA 0x55` are discarded.
2. **Bounded payload**: if `payload_len` exceeds host maximum, drop one byte and search for next sync.
3. **CRC validation**: if CRC fails, drop one byte and rescan for `0xAA 0x55`.
4. **Version/payload validation**: unknown versions or malformed payloads are treated as invalid frames and dropped.
5. **Partial buffering**: if not enough bytes are available for a full frame, keep buffered bytes and wait for more UART input.

Firmware behavior on error conditions:

- If a sensor read fails, firmware SHOULD still emit a frame and encode an out-of-range sentinel value per channel policy.
- Firmware MUST keep frame structure valid (correct length + CRC) even when a sensor value is invalid.
