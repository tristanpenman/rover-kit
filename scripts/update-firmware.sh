#!/bin/bash

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

TARGET=${TARGET:-stm32f4disco}

if command -v STM32_Programmer_CLI >/dev/null 2>&1; then
  echo "STM32_Programmer_CLI found in PATH."
  PROG=STM32_Programmer_CLI
else
  echo "STM32_Programmer_CLI not found in PATH. Checking common install locations..."
  if [[ "$OSTYPE" == "darwin"* ]]; then
    APP="/Applications/STMicroelectronics/STM32Cube/STM32CubeProgrammer/STM32CubeProgrammer.app"
    PROG="${APP}/Contents/Resources/bin/STM32_Programmer_CLI"
    if [[ -f "$PROG" ]]; then
      echo "STM32_Programmer_CLI found in: ${APP}"
    else
      unset PROG
    fi
  fi
fi

if [[ -z "$PROG" || ! -x "$PROG" ]]; then
  echo "Error: STM32_Programmer_CLI not found. Please install it and ensure it's in your PATH."
  exit 1
fi

${PROG} -c port=SWD -w bin/sonar-${TARGET}.elf 0x08000000
