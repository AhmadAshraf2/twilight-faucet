package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"github.com/rs/cors"
	"github.com/spf13/viper"
)

type RequestPayload struct {
	RecipientAddress string `json:"recipientAddress"`
}

var VALIDATOR_NAME string

var (
	dbHost     string
	dbPort     int
	dbUser     string
	dbPassword string
	dbName     string
)

func generateRandomHash() (string, error) {
	// Create a 32-byte array
	randomBytes := make([]byte, 32)

	// Fill the array with random bytes
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode the bytes into a hexadecimal string
	hash := hex.EncodeToString(randomBytes)
	return hash, nil
}

// runBTCDepositConfirmation runs the specified command with the provided parameters.
func runBTCDepositConfirmation(recipientAddress string) error {
	addrBytes, err := exec.Command(
		"nyksd", "keys", "show", VALIDATOR_NAME,
		"-a", "--keyring-backend", "test",
	).Output()
	if err != nil {
		return fmt.Errorf("failed to get validator address: %w", err)
	}
	validatorAddr := strings.TrimSpace(string(addrBytes)) // remove trailing newline

	tx_id, err := generateRandomHash()
	if err != nil {
		return fmt.Errorf("failed to generate random hash: %w", err)
	}
	cmd := exec.Command(
		"nyksd", "tx", "bridge", "msg-confirm-btc-deposit", "14uEN8abvKA1zgYCpv8MWCUwAMLGBqdZGM", viper.GetString("satsAmount"), "50000",
		tx_id,
		recipientAddress,
		validatorAddr,
		"--from", VALIDATOR_NAME,
		"--chain-id", "nyks",
		"--keyring-backend", "test",
		"--yes",
	)
	// 2. Force the right HOME so nyksd sees your test keyring
	// cmd.Env = append(os.Environ(),"HOME=${HOME}",)

	// Run the command and capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute command: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("Command executed successfully:\n%s\n", string(output))
	return nil
}

// runBTCDepositConfirmation runs the specified command with the provided parameters.
func runBTCDepositConfirmationRelayerWallet(recipientAddress string) error {
	addrBytes, err := exec.Command(
		"nyksd", "keys", "show", VALIDATOR_NAME,
		"-a", "--keyring-backend", "test",
	).Output()
	if err != nil {
		return fmt.Errorf("failed to get validator address: %w", err)
	}
	validatorAddr := strings.TrimSpace(string(addrBytes)) // remove trailing newline

	tx_id, err := generateRandomHash()
	if err != nil {
		return fmt.Errorf("failed to generate random hash: %w", err)
	}
	cmd := exec.Command(
		"nyksd", "tx", "bridge", "msg-confirm-btc-deposit", "14uEN8abvKA1zgYCpv8MWCUwAMLGBqdZGM", "500000000", "50000",
		tx_id,
		recipientAddress,
		validatorAddr,
		"--from", VALIDATOR_NAME,
		"--chain-id", "nyks",
		"--keyring-backend", "test",
		"--yes",
	)
	// 2. Force the right HOME so nyksd sees your test keyring
	// cmd.Env = append(os.Environ(),"HOME=${HOME}",)

	// Run the command and capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute command: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("Command executed successfully:\n%s\n", string(output))
	return nil
}

// =================================================================================================

func runBankSendCommand(toAddress string) error {
	// Construct the command
	cmd := exec.Command(
		"nyksd", "tx", "bank", "send",
		viper.GetString("faucet_account_name"),
		toAddress,
		viper.GetString("nyksAmount"),
		"--keyring-backend", "test",
		"--chain-id", "nyks",
		"--yes",
	)
	// 2. Force the right HOME so nyksd sees your test keyring
	// cmd.Env = append(os.Environ(),"HOME=${HOME}",)
	// Run the command and capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute command: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("Command executed successfully:\n%s\n", string(output))
	return nil
}

//================================================================================================

func handlemint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var payload RequestPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if payload.RecipientAddress == "" {
		http.Error(w, "recipientAddress is required", http.StatusBadRequest)
		return
	}

	exists, err := addressExists(payload.RecipientAddress)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error checking address existence: %v", err), http.StatusInternalServerError)
		return
	}
	if exists {
		used, err := btcRecentlyUsed(payload.RecipientAddress)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error checking address usage: %v", err), http.StatusInternalServerError)
			return
		}
		if used {
			http.Error(w, "Address has received btc in the last 24 hours", http.StatusTooManyRequests)
			return
		}
	}

	err = runBTCDepositConfirmation(payload.RecipientAddress)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to run command: %v", err), http.StatusInternalServerError)
		return
	}

	if !exists {
		err = insertAddressWithBtcTime(payload.RecipientAddress)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to insert address: %v", err), http.StatusInternalServerError)
			return
		}
	} else {
		err = updateBtcTime(payload.RecipientAddress)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to update address time: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Command executed successfully"))
}

//================================================================================================

func handlemintRelayerWallet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var payload RequestPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if payload.RecipientAddress == "" {
		http.Error(w, "recipientAddress is required", http.StatusBadRequest)
		return
	}

	if payload.RecipientAddress == viper.GetString("relayer_address") {
		http.Error(w, "recipientAddress is not the register relayer address", http.StatusBadRequest)
		return
	}

	err = runBTCDepositConfirmationRelayerWallet(payload.RecipientAddress)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to run command: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Command executed successfully"))
}

//================================================================================================

func handlefaucet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var payload RequestPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if payload.RecipientAddress == "" {
		http.Error(w, "recipientAddress is required", http.StatusBadRequest)
		return
	}

	exists, err := addressExists(payload.RecipientAddress)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error checking address existence: %v", err), http.StatusInternalServerError)
		return
	}
	if exists {
		used, err := nyksRecentlyUsed(payload.RecipientAddress)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error checking address usage: %v", err), http.StatusInternalServerError)
			return
		}
		if used {
			http.Error(w, "Address has received nyks in the last 24 hours", http.StatusTooManyRequests)
			return
		}
	}

	err = runBankSendCommand(payload.RecipientAddress)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to run command: %v", err), http.StatusInternalServerError)
		return
	}

	if !exists {
		err = insertAddressWithNyksTime(payload.RecipientAddress)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to insert address: %v", err), http.StatusInternalServerError)
			return
		}
	} else {
		err = updateNyksTime(payload.RecipientAddress)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to update address time: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Command executed successfully"))
}

//================================================================================================

func main() {
	initialize()
	mux := http.NewServeMux()
	mux.HandleFunc("/mint", handlemint)
	mux.HandleFunc("/faucet", handlefaucet)
	mux.HandleFunc("/mint-relayer-wallet", handlemintRelayerWallet)
	// Configure
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	})

	handler := c.Handler(mux)
	fmt.Println("Server is running on port 6969  with CORS...")
	err := http.ListenAndServe(":6969", handler)
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
	// log.Fatal(http.ListenAndServe(":6969", handler))
}

func initialize() {
	viper.AddConfigPath("./config")
	viper.SetConfigName("config") // Register config file name (no extension)
	viper.SetConfigType("json")   // Look for specific type
	viper.ReadInConfig()

	VALIDATOR_NAME = viper.GetString("validator_name")
	dbHost = viper.GetString("db_host")
	dbPort = viper.GetInt("db_port")
	dbUser = viper.GetString("db_user")
	dbPassword = viper.GetString("db_password")
	dbName = viper.GetString("db_name")
}
