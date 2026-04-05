# Plan

* Refactoring
  * ~~Port web bridge code to Go~~
  * ~~Dockerfile or build scripts for ARMv6~~
  * ~~Implement motor driver using Gobot~~
  * ~~Implement motor driver using periph library~~
  * Modularise sonar reader (**in progress**)
  * Implement periph provider for sonar

* Tests
  * Test harness for motor and sonar modules
    * Parser and validation tests for motor payloads 
    * Parser and validation tests for sonar payloads 
  * Integration tests with a mocked MQTT broker to validate publish/subscribe flow

* Microcontroller
  * Basic structure for STM32 firmware
  * Explore using TinyGo or Embedded Go
  * Define UART protocol for sonar readings
  * Full STM32 based sonar receiver

* Future
  * Improve web interface
  * STL files for a new chassis
  * Investigate better motors
  * Add a camera module
