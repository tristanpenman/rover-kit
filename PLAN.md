# Plan

* Refactoring
  * ~~Port web bridge code to Go~~
  * ~~Dockerfile or build scripts for ARMv6~~
  * ~~Implement motor driver using Gobot~~
  * ~~Implement motor driver using periph library~~
  * ~~Modularise sonar reader~~
  * Implement periph provider for sonar (**in progress**)

* Tests
  * Test harness for motor and sonar modules (**in progress**)
    * Parser and validation tests for motor payloads
    * Parser and validation tests for sonar payloads
  * Integration tests
    * Mocked MQTT broker to validate publish/subscribe flow

* Microcontroller
  * ~~Basic structure for STM32 firmware~~
  * Explore using TinyGo or Embedded Go (**in progress**)
  * Define UART protocol for sonar readings
  * Full STM32 based sonar receiver

* Future
  * Improve web interface
  * STL files for a new chassis
  * Investigate better motors
  * Add a camera module
