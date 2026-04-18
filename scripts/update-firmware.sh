#!/bin/bash

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VARIATION=${1:-sonar}
TARGET=${TARGET:-stm32f4disco}

if command -v STM32_Programmer_CLI >/dev/null 2>&1; then
  echo "STM32_Programmer_CLI found in PATH."
  CLI=$(which STM32_Programmer_CLI)
else
  echo "STM32_Programmer_CLI not found in PATH. Checking common install locations..."

  if [[ "$OSTYPE" == "darwin"* ]]; then
    # Check for STM32CubeProgrammer.app on macOS
    APP="/Applications/STMicroelectronics/STM32Cube/STM32CubeProgrammer/STM32CubeProgrammer.app"
    DIR="$APP/Contents/Resources/bin"
    if [[ -f "${DIR}/STM32_Programmer_CLI" ]]; then
      CLI="${DIR}/STM32_Programmer_CLI"
      echo "STM32_Programmer_CLI found in: ${DIR}"
    fi
  else
    # Check common Linux install locations
    for DIR in /usr/local/bin /usr/bin /opt/stm32cubeprogrammer/bin ${HOME}/STMicroelectronics/STM32Cube/STM32CubeProgrammer/bin; do
      if [[ -x "${DIR}/STM32_Programmer_CLI" ]]; then
        CLI="${DIR}/STM32_Programmer_CLI"
        echo "STM32_Programmer_CLI found in: ${DIR}"
        break
      fi
    done
  fi
fi

if [[ -z "$CLI" || ! -x "$CLI" ]]; then
  echo "Error: STM32_Programmer_CLI not found. Please install it and ensure it's in your PATH."
  exit 1
fi

sudo ${CLI} -c port=SWD -w bin/${VARIATION}-${TARGET}.bin 0x08000000
