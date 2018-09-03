package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ATCreateCommandContext Create AT per command context
func ATCreateCommandContext(parentCtx context.Context) (ctx context.Context, cancel context.CancelFunc) {

	duration := 5 * time.Second

	// Allow more time to proceed when not executing on real device with debug
	if !IsTargetDevice() {
		duration = 60 * time.Second
	}

	ctx, cancel = context.WithTimeout(parentCtx, duration)
	return ctx, cancel
}

func handleAT(ctx context.Context, port *Port, timeout time.Duration, logger *Logger, modemStatusMessageChannel chan ModemStatusMessage) (bool, error) {

	// Global initing for this session
	timeoutReader := NewReader(port, timeout)
	reader := bufio.NewReader(timeoutReader)
	handler := &AtCommandHandler{logger: logger,
		reader: reader,
		writer: port}

	errorModeTextEnabled := false
	initialConnected := true

	for {
		select {
		case <-ctx.Done():
			if IsDebugMode() {
				logger.Debug("Cancelled modem command handling")
			}
			return false, nil
		default:
			modemAvailable := true
			simpinOk := true
			gotSimID := true
			signal := NoSignal
			connType := ConnTypeNoNetwork
			csq := 0
			ber := 0

			if IsDebugMode() {
				logger.Debugf("Trying to fetch modem data")
			}

			// First try simple at command.
			err := AT(ctx, handler)

			if err := TryHandleAtCommandError(logger, "AT", err, func() { modemAvailable = false; initialConnected = false }); err != nil {
				return initialConnected, err
			}

			// Disable AT echo
			err = ATE(ctx, handler, false)
			if err := TryHandleAtCommandError(logger, "ATE", err, func() { modemAvailable = false }); err != nil {
				logger.Warning("ATE failed")
			}

			// Then try to enable modem errors
			if !errorModeTextEnabled {
				if ATCMEE(ctx, handler, 2) != nil {
					errorModeTextEnabled = true
				}
			}

			// Check for SIM and PIN
			err = ATCPIN(ctx, handler)

			if err := TryHandleAtCommandError(logger, "AT+CPIN?", err, func() { simpinOk = false }); err != nil {
				return initialConnected, err
			}

			// Check signal quality
			csqRes, err := ATCSQ(ctx, handler)

			if err := TryHandleAtCommandError(logger, "AT+CSQ", err, func() { signal = NoSignal }); err != nil {
				return initialConnected, err
			}

			// Copy the result values back
			if err == nil {
				csq = csqRes.Csq
				ber = csqRes.Ber
			}

			// Check broadband connection type
			connType, err = ATCNSMOD(ctx, handler)

			// Check modem connection type
			if err := TryHandleAtCommandError(logger, "AT+CNSMOD?", err, func() { connType = ConnTypeNoNetwork }); err != nil {
				return initialConnected, err
			}

			// GET SIM ID
			str, err := ATCCID(ctx, handler)

			if err := TryHandleAtCommandError(logger, "AT+CCID", err, func() { gotSimID = true }); err != nil {
				return initialConnected, err
			}

			signal = TranslateModemDBM(csq, ber)

			modemStatusMessageChannel <- ModemStatusMessage{
				ModemAvailable:    modemAvailable,
				DataAvailable:     simpinOk,
				SignalStrength:    signal,
				SimpinOk:          simpinOk,
				SimUccid:          str,
				BroadbandConnType: connType,
			}
		}
	}
}

// TryHandleAtCommandError try handle a command error return false if it's not being handled
func TryHandleAtCommandError(logger *Logger, cmd string, err error, atErrorHandler func()) error {

	if err == nil {
		return nil
	}

	if IsDebugMode() {
		logger.Debugf("Error: %v in command: %v", err.Error(), cmd)
	}

	// Check if we can cast to an error from which we can recover
	if _, ok := err.(*SimError); ok {

		if ok {

			// log in debug mode
			if IsDebugMode() {
				logger.Debugf("Ignoring error: %v in command: %v", err.Error(), cmd)
			}

			// Call the handler
			atErrorHandler()

			return nil
		}

		return err
	}
	// fallback
	return err
}

// AtCommandHandler structure
type AtCommandHandler struct {
	reader *bufio.Reader
	writer *Port
	logger *Logger
}

// HandleCommand the serial handler
func (atCommandHandler *AtCommandHandler) HandleCommand(parentCtx context.Context, f func(ctx context.Context, cancel context.CancelFunc) error) error {

	commandCtx, cancel := ATCreateCommandContext(parentCtx)

	// Defer cancellation.
	defer cancel()

	return f(commandCtx, cancel)
}

// HandleCommandWithOutput the serial handler
func (atCommandHandler *AtCommandHandler) HandleCommandWithOutput(parentCtx context.Context, f func(ctx context.Context, cancel context.CancelFunc) (d interface{}, e error)) (data interface{}, err error) {

	// Drain the serial port before executing
	for {

		str, err := atCommandHandler.reader.ReadString('\r')
		str = strings.TrimSpace(str)
		if str != "" && IsDebugMode() {
			atCommandHandler.logger.DebugF("Drained following garbage: %v from serial port", str)
		}

		if err != nil {
			break
		}
	}

	commandCtx, cancel := ATCreateCommandContext(parentCtx)
	return f(commandCtx, cancel)
}

// AtHandle Structure
type AtHandle struct {
	Command string
	ctx     context.Context
	cancel  context.CancelFunc
	handler func(line string) (completed bool, flow bool, err error)
}

// ErrCommandCancelled the command is cancelled
var ErrCommandCancelled = errors.New("Cancelled")

// Execute the at command
func (atCommand *AtHandle) Execute(handler *AtCommandHandler) error {

	cmd := atCommand.Command + "\r"
	cmdBytes := []byte(cmd)

	n, err := handler.writer.Write(cmdBytes)

	if err != nil {
		return err
	}

	if n != len(cmdBytes) {
		return errors.New("invallid data length")
	}

	// Try to sleep a while
	time.Sleep(500 * time.Millisecond)

	str, err := handler.reader.ReadString('\r')
	str = strings.TrimSpace(str)

	for {

		select {
		case <-atCommand.ctx.Done():

			if IsDebugMode() {
				fmt.Printf("timing failure in command: %v\r\n", atCommand.Command)
			}

			return ErrCommandCancelled
		default:

			if err != nil {
				return err
			}

			completed, continueFlow, err := atCommand.handler(str)

			if err != nil && completed {
				return err
			}

			if !continueFlow {
				return nil
			}

			str, err = handler.reader.ReadString('\r')
			str = strings.TrimSpace(str)
		}
	}
}

// AT command function
func AT(parentCtx context.Context, handler *AtCommandHandler) error {

	return handler.HandleCommand(parentCtx, func(ctx context.Context, cancel context.CancelFunc) error {

		command := &AtHandle{
			Command: "AT",
			ctx:     ctx,
			cancel:  cancel,
			handler: DefaultATHandler(),
		}

		return command.Execute(handler)
	})
}

// ATE command function
func ATE(parentCtx context.Context, handler *AtCommandHandler, enableOrDisableEcho bool) error {

	return handler.HandleCommand(parentCtx, func(ctx context.Context, cancel context.CancelFunc) error {

		cmd := ""

		if enableOrDisableEcho {
			cmd = "ATE1"
		} else {
			cmd = "ATE0"
		}

		command := &AtHandle{
			Command: cmd,
			ctx:     ctx,
			cancel:  cancel,
			handler: DefaultATHandler(),
		}

		return command.Execute(handler)
	})
}

// ATCMEE command function
func ATCMEE(parentCtx context.Context, handler *AtCommandHandler, level int) error {

	return handler.HandleCommand(parentCtx, func(ctx context.Context, cancel context.CancelFunc) error {
		cmd := fmt.Sprintf("AT+CMEE=%v", level)

		command := &AtHandle{
			Command: cmd,
			ctx:     ctx,
			cancel:  cancel,
			handler: DefaultATHandler(),
		}

		return command.Execute(handler)
	})
}

//CsqResult type
type CsqResult struct {
	Csq int
	Ber int
}

// ATCSQ command function
func ATCSQ(parentCtx context.Context, handler *AtCommandHandler) (csqRes CsqResult, err error) {

	res, err := handler.HandleCommandWithOutput(parentCtx, func(ctx context.Context, cancel context.CancelFunc) (csq interface{}, err error) {

		csqValue := 0
		berValue := 0

		command := &AtHandle{
			Command: "AT+CSQ",
			ctx:     ctx,
			cancel:  cancel,
			handler: ATPrefixHandler("+CSQ:", func(line string) (bool, bool, error) {

				items := strings.Split(line, ",")
				csqStr := strings.Replace(items[0], "+CSQ: ", "", -1)
				csqValue, _ = strconv.Atoi(csqStr)
				berValue, _ = strconv.Atoi(items[1])

				return ATCompletedReadNext()
			})}

		err = command.Execute(handler)
		res := CsqResult{
			Csq: csqValue,
			Ber: berValue}

		return res, err
	})

	// Type cast magic
	csq, ok := res.(CsqResult)

	if ok {
		return csq, err
	}

	return CsqResult{}, err

}

// ATCNSMOD command function
func ATCNSMOD(parentCtx context.Context, handler *AtCommandHandler) (bct BroadbandConnType, err error) {

	res, err := handler.HandleCommandWithOutput(parentCtx, func(ctx context.Context, cancel context.CancelFunc) (ct interface{}, err error) {

		ct = ConnTypeNoNetwork
		command := &AtHandle{
			Command: "AT+CNSMOD?",
			ctx:     ctx,
			cancel:  cancel,
			handler: ATPrefixHandler("+CNSMOD", func(line string) (bool, bool, error) {

				items := strings.Split(line, ",")
				n, err := strconv.Atoi(items[1])

				if err == nil && n > 0 {

					if n <= 3 {
						ct = ConnType2G
					} else {
						ct = ConnType3G
					}
					// Todo: Add 4G support
				}

				return ATCompletedReadNext()
			})}

		err = command.Execute(handler)

		return ct, err
	})

	// Type cast magic
	bctr, ok := res.(BroadbandConnType)

	if ok {
		return bctr, err
	}

	return ConnTypeNoNetwork, err
}

func getCcidFromCcidLine(line string) string {

	items := strings.Split(line, ",")
	ccIDStr := strings.Replace(items[0], "+CCID: ", "", -1)

	return strings.TrimRight(strings.TrimLeft(ccIDStr, "\""), "\"")
}

// ATCCID command function
func ATCCID(parentCtx context.Context, handler *AtCommandHandler) (ccid string, err error) {

	res, err := handler.HandleCommandWithOutput(parentCtx, func(ctx context.Context, cancel context.CancelFunc) (res interface{}, err error) {

		command := &AtHandle{
			Command: "AT+CCID",
			ctx:     ctx,
			cancel:  cancel,
			handler: ATPrefixHandler("+CCID:", func(line string) (bool, bool, error) {

				ccid = getCcidFromCcidLine(line)
				return ATCompletedReadNext()
			})}

		err = command.Execute(handler)
		return ccid, err
	})

	result, ok := res.(string)
	if ok {
		return result, nil
	}

	return "", errors.New("cast failure")
}

// ATCPIN command function
func ATCPIN(parentCtx context.Context, handler *AtCommandHandler) error {

	return handler.HandleCommand(parentCtx, func(ctx context.Context, cancel context.CancelFunc) error {

		cmd := "AT+CPIN?"

		command := &AtHandle{
			Command: cmd,
			ctx:     ctx,
			cancel:  cancel,
			handler: ATPrefixHandler("+CPIN:", func(line string) (bool, bool, error) {

				line = strings.Replace(line, "+CPIN: ", "", -1)
				line = strings.ToUpper(line)

				switch line {

				case "READY":
					return ATReadNextLine()
				case "SIM PIN":
					return ATErrorNext(ErrorFromSimState(PinLocked), true)
				case "SIM PUK":
					return ATErrorNext(ErrorFromSimState(PinLocked), true)
				case "SIM PIN2":
					return ATErrorNext(ErrorFromSimState(PinLocked2), true)
				case "SIM PUK2":
					return ATErrorNext(ErrorFromSimState(PukLocked2), true)
				default:
					return ATErrorNext(ErrorFromSimState(UnkownState), true)
				}
			})}

		return command.Execute(handler)
	})
}

// ATReadNextLine we want more data
func ATReadNextLine() (completed bool, flow bool, err error) {
	return false, true, nil
}

// ATCheckOk checks for OK messages
func ATCheckOk(line string) (flow bool) {

	// Check if we're ok
	if line == "OK" {
		return true
	}

	return false

}

// DefaultATHandler use default at handling mechanism
func DefaultATHandler() func(line string) (completed bool, flow bool, err error) {

	return func(line string) (completed bool, flow bool, err error) {

		err = DefaultATErrorHandler(line)

		// First run the default error handler
		if err != nil {
			return ATError(err)
		}

		// Check if we're ok
		if ATCheckOk(line) {
			return ATCompleted()
		}

		//Skip a line and check again
		return ATReadNextLine()
	}
}

// ATPrefixHandler use default at handling mechanism
func ATPrefixHandler(prefix string, prefixFunc func(line string) (completed bool, flow bool, err error)) func(line string) (completed bool, flow bool, err error) {

	return func(line string) (completed bool, flow bool, err error) {

		err = DefaultATErrorHandler(line)

		if err != nil {
			return ATError(err)
		}

		// Check if we're matching prefix
		if strings.HasPrefix(line, prefix) {
			return prefixFunc(line)
		}

		// Check if we're ok
		if ATCheckOk(line) {
			return ATCompleted()
		}

		//Skip a line and check again
		return ATCompletedReadNext()
	}
}

// DefaultATErrorHandler default logic for error handling
func DefaultATErrorHandler(line string) error {

	if strings.HasPrefix(line, "ERROR") {
		return ErrorFromATText("unspecified error (try to set error format for more information)")
	}

	if strings.HasPrefix(line, "+CME ERROR: ") {
		text := strings.Replace(line, "+CME ERROR: ", "", -1)
		return ErrorFromATText(text)
	}

	return nil
}

// ATCompleted we are completed and expect no more data
func ATCompleted() (completed bool, flow bool, err error) {
	return true, false, nil
}

// ATCompletedReadNext we parsed as completed but we expect more commands
func ATCompletedReadNext() (completed bool, flow bool, err error) {
	return true, true, nil
}

// ATError we are completed but have an error
func ATError(cause error) (completed bool, flow bool, err error) {
	return true, false, cause
}

// ATErrorNext we are completed but have an error and expect more data
func ATErrorNext(cause error, next bool) (completed bool, flow bool, err error) {
	return true, next, cause
}

// SimErrorState value
type SimErrorState uint

const (
	// UnkownState means unkown
	UnkownState SimErrorState = 0
	// NoSim means unkown
	NoSim SimErrorState = 1
	// PinLocked means locked by pin code
	PinLocked SimErrorState = 2
	// PinLocked2 means locked by pin code
	PinLocked2 SimErrorState = 3
	// PukLocked means locked by pin code
	PukLocked SimErrorState = 4
	// PukLocked2 means locked by pin code
	PukLocked2 SimErrorState = 5
)

// SimError custom sim error type
type SimError struct {
	error string
}

// ErrorFromSimState creates error from sim state
func ErrorFromSimState(errorState SimErrorState) error {
	return &SimError{error: fmt.Sprintf("Got sim error: %v", errorState)}
}

// ErrorFromATText creates error from sim state
func ErrorFromATText(err string) error {
	return &SimError{error: err}
}

func (simError *SimError) Error() string {
	return simError.error
}
