# Plan

* Refactoring
  * ~~Port web bridge code to Go~~
  * ~~Dockerfile or build scripts for ARMv6~~
  * ~~Implement motor driver using Gobot~~
  * Implement motor driver using periph library (**in progress**)
  * Modularise sonar reader

* Tests
  * Test harness for motor and sonar modules
    * Parser and validation tests for motor payloads 
    * Parser and validation tests for sonar payloads 
  * Integration tests with a mocked MQTT broker to validate publish/subscribe flow

* Microcontroller
  * Basic structure for STM32 firmware
  * Explore using TinyGo or Embedded Go
  * Full STM32 based sonar receiver

* Future
  * Improve web interface
  * STL files for a new chassis
  * Investigate better motors
  * Add a camera module
