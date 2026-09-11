package cli

import (
	"database/sql"
	"fmt"

	"github.com/ergochat/readline"
	"github.com/pquerna/otp/totp"

	"github.com/joshu-sajeev/secure-cli/internal/auth"
)

// Enable2FA handles the enable-2fa command
func Enable2FA(db *sql.DB, rl *readline.Instance, userID int64, username string) error {
	userDetails, err := auth.GetUserDetails(db, userID)
	if err != nil {
		fmt.Fprintln(rl, "❌ Error retrieving user information.")
		return err
	}

	if userDetails.MFAEnabled {
		fmt.Fprintln(rl, "⚠️  Two-Factor Authentication is already enabled for your account.")
		return nil
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Secure-CLI",
		AccountName: username,
	})
	if err != nil {
		fmt.Fprintln(rl, "❌ Failed to generate 2FA secret.")
		return err
	}

	fmt.Fprintln(rl, "")
	fmt.Fprintln(rl, "═════════════════════════════════════════════════════════════")
	fmt.Fprintln(rl, "  📱 Two-Factor Authentication Setup")
	fmt.Fprintln(rl, "═════════════════════════════════════════════════════════════")
	fmt.Fprintln(rl, "")

	fmt.Fprintln(rl, "STEP 1: Add your account to an authenticator app")
	fmt.Fprintln(rl, "")
	fmt.Fprintln(rl, "Open Google Authenticator, Authy, Microsoft Authenticator,")
	fmt.Fprintln(rl, "or another TOTP-compatible authenticator app.")
	fmt.Fprintln(rl, "")
	fmt.Fprintln(rl, "Choose 'Enter setup key' or 'Enter key manually'.")
	fmt.Fprintln(rl, "")

	fmt.Fprintln(rl, "Account Name:")
	fmt.Fprintf(rl, "  Secure-CLI (%s)\n", username)
	fmt.Fprintln(rl, "")

	fmt.Fprintln(rl, "Manual Entry Key:")
	fmt.Fprintf(rl, "  🔑 %s\n", key.Secret())
	fmt.Fprintln(rl, "")

	fmt.Fprintln(rl, "STEP 2: Verify Your Setup")
	fmt.Fprintln(rl, "Enter the 6-digit code from your authenticator:")
	fmt.Fprintln(rl, "")

	fmt.Fprint(rl, "2FA Code: ")
	verificationCode, _ := rl.ReadLine()

	if !totp.Validate(verificationCode, key.Secret()) {
		fmt.Fprintln(rl, "")
		fmt.Fprintln(rl, "❌ Invalid code. The code may have expired. Please try again.")
		return fmt.Errorf("invalid 2fa verification code")
	}

	if err := enableTOTPInDatabase(db, userID, key.Secret()); err != nil {
		fmt.Fprintln(rl, "❌ Failed to save 2FA settings to database.")
		return err
	}

	fmt.Fprintln(rl, "")
	fmt.Fprintln(rl, "✅ Two-Factor Authentication successfully enabled!")
	fmt.Fprintln(rl, "")
	fmt.Fprintln(rl, "⚠️  IMPORTANT: Save Your Secret Key")
	fmt.Fprintln(rl, "───────────────────────────────────────")
	fmt.Fprintf(rl, "Secret Key: %s\n", key.Secret())
	fmt.Fprintln(rl, "")
	fmt.Fprintln(rl, "⚠️  If you lose access to your authenticator app, you can use")
	fmt.Fprintln(rl, "   this secret key to recover your account (contact admin).")
	fmt.Fprintln(rl, "")
	fmt.Fprintln(rl, "From now on, you'll need to enter a 6-digit code when logging in.")
	fmt.Fprintln(rl, "═════════════════════════════════════════════════════════════")
	fmt.Fprintln(rl, "")

	return nil
}

// Disable2FA handles the disable-2fa command
func Disable2FA(db *sql.DB, rl *readline.Instance, userID int64, username string) error {
	userDetails, err := auth.GetUserDetails(db, userID)
	if err != nil {
		fmt.Fprintln(rl, "❌ Error retrieving user information.")
		return err
	}

	if !userDetails.MFAEnabled {
		fmt.Fprintln(rl, "⚠️  Two-Factor Authentication is not enabled for your account.")
		return nil
	}

	fmt.Fprintln(rl, "")
	fmt.Fprintln(rl, "⚠️  Disabling 2FA requires password confirmation for security.")
	fmt.Fprintln(rl, "")
	fmt.Fprint(rl, "Enter your password: ")
	password, _ := rl.ReadLine()

	_, err = auth.Login(db, username, password)
	if err != nil {
		fmt.Fprintln(rl, "")
		fmt.Fprintln(rl, "❌ Incorrect password. Cannot disable 2FA.")
		return err
	}

	fmt.Fprintln(rl, "")
	fmt.Fprintln(rl, "⚠️  WARNING:")
	fmt.Fprintln(rl, "   Disabling 2FA will make your account less secure.")
	fmt.Fprintln(rl, "   Are you sure you want to continue?")
	fmt.Fprintln(rl, "")
	fmt.Fprint(rl, "Type 'yes' to confirm: ")
	confirmation, _ := rl.ReadLine()

	if confirmation != "yes" {
		fmt.Fprintln(rl, "Cancelled. 2FA remains enabled.")
		return nil
	}

	if err := disableTOTPInDatabase(db, userID); err != nil {
		fmt.Fprintln(rl, "❌ Failed to disable 2FA.")
		return err
	}

	fmt.Fprintln(rl, "")
	fmt.Fprintln(rl, "✅ Two-Factor Authentication successfully disabled.")
	fmt.Fprintln(rl, "")

	return nil
}

// VerifyTOTPDuringLogin prompts for and verifies TOTP code during login
func VerifyTOTPDuringLogin(rl *readline.Instance, totpSecret string) (bool, error) {
	fmt.Fprintln(rl, "")
	fmt.Fprintln(rl, "🔐 Two-Factor Authentication Required")
	fmt.Fprintln(rl, "Enter the 6-digit code from your authenticator app:")
	fmt.Fprintln(rl, "")

	maxAttempts := 3
	for i := 0; i < maxAttempts; i++ {
		fmt.Fprintf(rl, "2FA Code (attempt %d/%d): ", i+1, maxAttempts)
		code, _ := rl.ReadLine()

		if totp.Validate(code, totpSecret) {
			fmt.Fprintln(rl, "✅ 2FA verification successful!")
			return true, nil
		}

		remaining := maxAttempts - i - 1
		if remaining > 0 {
			fmt.Fprintf(rl, "❌ Invalid code. %d attempts remaining.\n", remaining)
		} else {
			fmt.Fprintln(rl, "❌ Too many failed attempts. Login cancelled.")
		}
	}

	return false, fmt.Errorf("2fa verification failed: too many attempts")
}

// Status2FA displays the 2FA status of the current user
func Status2FA(db *sql.DB, rl *readline.Instance, userID int64) error {
	userDetails, err := auth.GetUserDetails(db, userID)
	if err != nil {
		fmt.Fprintln(rl, "❌ Error retrieving user information.")
		return err
	}

	fmt.Fprintln(rl, "")
	fmt.Fprintln(rl, "───────────────────────────────")
	if userDetails.MFAEnabled {
		fmt.Fprintln(rl, "✅ Two-Factor Authentication: ENABLED")
	} else {
		fmt.Fprintln(rl, "❌ Two-Factor Authentication: DISABLED")
	}
	fmt.Fprintln(rl, "───────────────────────────────")
	fmt.Fprintln(rl, "")

	return nil
}

// enableTOTPInDatabase updates the user record to enable TOTP
func enableTOTPInDatabase(db *sql.DB, userID int64, secret string) error {
	_, err := db.Exec(
		"UPDATE users SET totp_secret = ?, totp_enabled = 1 WHERE id = ?",
		secret,
		userID,
	)
	return err
}

// disableTOTPInDatabase updates the user record to disable TOTP
func disableTOTPInDatabase(db *sql.DB, userID int64) error {
	_, err := db.Exec(
		"UPDATE users SET totp_secret = NULL, totp_enabled = 0 WHERE id = ?",
		userID,
	)
	return err
}
