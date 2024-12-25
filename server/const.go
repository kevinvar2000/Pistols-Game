package main

// Constants for the server connection
const (
	CONN_ADDRESS = "0.0.0.0" // Address to bind the server
	CONN_PORT    = "8080"    // Port to bind the server
	CONN_NETWORK = "tcp"     // Network protocol to use
)

// Constants for the game settings
const (
	MAX_PLAYERS           = 2        // Maximum number of players in the game
	DEFAULT_HEALTH        = 3        // Default health for each player
	DEFAULT_AMMO          = 1        // Default ammo for each player
	DEFAULT_ACTION        = "COVER"  // Default action for each player
	ACTION_SHOOT          = "SHOOT"  // Action to shoot the opponent
	ACTION_COVER          = "COVER"  // Action to cover and protect oneself
	ACTION_RELOAD         = "RELOAD" // Action to reload the weapon
	ACTION_TIMEOUT        = 20       // Timeout for player actions in seconds
	ACTION_CHECK_INTERVAL = 500      // Interval to check player actions in milliseconds
	ROUND_PAUSE_TIME      = 5        // Pause time between rounds in seconds
	READ_TIMEOUT          = 20       // Time interval to ping clients in seconds
	READ_DEADLINE         = 30       // Deadline for reading from clients in seconds
)
