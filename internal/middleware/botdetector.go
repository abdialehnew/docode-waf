package middleware

import (
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

var badBotPatterns = []string{
	"(?i)(bot|crawler|spider|scraper)",
	"(?i)(wget|curl|python-requests)",
	"(?i)(semrush|ahrefs|majestic)",
}

// BotDetectorMiddleware detects and blocks known bad bots based on vhost settings
func BotDetectorMiddleware(db *sqlx.DB) gin.HandlerFunc {
	compiledPatterns := make([]*regexp.Regexp, 0, len(badBotPatterns))
	for _, pattern := range badBotPatterns {
		re, err := regexp.Compile(pattern)
		if err == nil {
			compiledPatterns = append(compiledPatterns, re)
		}
	}

	return func(c *gin.Context) {
		// Get current vhost domain
		domain := c.Request.Host
		if colonIdx := strings.Index(domain, ":"); colonIdx != -1 {
			domain = domain[:colonIdx]
		}

		// Check if bot detection is enabled for this vhost
		var vhostSettings struct {
			BotDetectionEnabled bool   `db:"bot_detection_enabled"`
			BotDetectionType    string `db:"bot_detection_type"`
			RecaptchaVersion    string `db:"recaptcha_version"`
		}

		err := db.Get(&vhostSettings, "SELECT bot_detection_enabled, bot_detection_type, recaptcha_version FROM vhosts WHERE domain = $1", domain)
		if err != nil {
			log.Printf("[Bot Detector] Error getting vhost settings for domain %s: %v", domain, err)
			c.Next()
			return
		}

		// If bot detection is disabled for this vhost, skip
		if !vhostSettings.BotDetectionEnabled {
			c.Next()
			return
		}

		// Check if user has already passed bot detection (cookie/session)
		if passed, _ := c.Cookie("bot_check_passed_" + domain); passed == "true" {
			c.Next()
			return
		}

		// Bot detection is enabled - show challenge to all visitors
		// They must complete the challenge before accessing the site
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusForbidden, getBotChallengeHTML(domain, vhostSettings.BotDetectionType, vhostSettings.RecaptchaVersion))
		c.Abort()
		return
	}
}

func isLegitimateBot(userAgent string) bool {
	legitimateBots := []string{
		"Googlebot",
		"Bingbot",
		"Slurp", // Yahoo
		"DuckDuckBot",
		"Baiduspider",
		"YandexBot",
		"facebookexternalhit",
		"LinkedInBot",
		"Twitterbot",
	}

	userAgentLower := strings.ToLower(userAgent)
	for _, bot := range legitimateBots {
		if strings.Contains(userAgentLower, strings.ToLower(bot)) {
			return true
		}
	}
	return false
}

// getBotChallengeHTML returns HTML for bot detection challenge
func getBotChallengeHTML(domain, challengeType, recaptchaVersion string) string {
	// Log the challenge type for debugging
	log.Printf("[Bot Detector] Generating challenge for domain %s with type: '%s'", domain, challengeType)

	var challengeContent, challengeTitle, challengeSubtitle string

	switch challengeType {
	case "turnstile":
		challengeTitle = "Cloudflare Turnstile Verification"
		challengeSubtitle = "Complete the challenge below to verify you're human"

		// Get Turnstile site key from environment
		siteKey := os.Getenv("TURNSTILE_SITE_KEY")
		if siteKey == "" || siteKey == "${TURNSTILE_SITE_KEY}" {
			siteKey = "0x4AAAAAAABdGKRzJgJuLBaZ" // Fallback to demo key
		}

		challengeContent = `
			<div class="turnstile-wrapper" style="display: flex; justify-content: center; align-items: center; min-height: 100px; margin: 30px 0;">
				<div id="turnstile-container"></div>
			</div>
			<script src="https://challenges.cloudflare.com/turnstile/v0/api.js?onload=onloadTurnstileCallback" defer></script>
			<script>
				// Callback when Turnstile script is loaded
				window.onloadTurnstileCallback = function() {
					console.log('Turnstile API loaded');
					try {
						turnstile.render('#turnstile-container', {
							sitekey: '` + siteKey + `',
							theme: 'auto',
							size: 'flexible',
							callback: function(token) {
								console.log('Turnstile verification successful');
								// Set cookie to mark bot check as passed
								document.cookie = "bot_check_passed_` + domain + `=true; path=/; max-age=3600; SameSite=Lax";
								
								// Show success message
								const container = document.querySelector('.container');
								const successMsg = document.createElement('div');
								successMsg.style.cssText = 'margin-top: 20px; padding: 15px; background: #d4edda; border: 1px solid #c3e6cb; border-radius: 8px; color: #155724; animation: slideIn 0.3s ease-out;';
								successMsg.innerHTML = '<strong>✅ Verification successful!</strong><br>Redirecting to application...';
								container.appendChild(successMsg);
								
								// Reload page after short delay
								setTimeout(() => {
									window.location.reload();
								}, 1500);
							},
							'error-callback': function(error) {
								console.error('Turnstile error:', error);
								const container = document.querySelector('.turnstile-wrapper');
								container.innerHTML = '<div style="color: #e53e3e; padding: 15px; background: #fff5f5; border: 1px solid #feb2b2; border-radius: 8px;">⚠️ Failed to load verification widget. Please refresh the page.</div>';
							},
							'expired-callback': function() {
								console.log('Turnstile token expired');
								window.location.reload();
							}
						});
					} catch (error) {
						console.error('Turnstile render error:', error);
						const container = document.querySelector('.turnstile-wrapper');
						container.innerHTML = '<div style="color: #e53e3e; padding: 15px; background: #fff5f5; border: 1px solid #feb2b2; border-radius: 8px;">⚠️ Verification widget failed to initialize. Please refresh the page.</div>';
					}
				};
				
				// Fallback: If script doesn't load after 10 seconds
				setTimeout(function() {
					if (typeof turnstile === 'undefined') {
						console.error('Turnstile script failed to load');
						const container = document.querySelector('.turnstile-wrapper');
						container.innerHTML = '<div style="color: #e53e3e; padding: 15px; background: #fff5f5; border: 1px solid #feb2b2; border-radius: 8px; text-align: center;"><strong>⚠️ Connection Error</strong><br><br>Could not load verification service from Cloudflare.<br>Please check your internet connection and refresh the page.</div>';
					}
				}, 10000);
			</script>
		`
	case "captcha":
		// Use recaptchaVersion parameter (default to v2 if empty)
		if recaptchaVersion == "" {
			recaptchaVersion = "v2"
		}

		if recaptchaVersion == "v3" {
			// reCAPTCHA v3 - Score-based, invisible
			challengeTitle = "Security Check"
			challengeSubtitle = "Verifying your browser security..."

			// Get reCAPTCHA v3 site key from environment
			recaptchaV3SiteKey := os.Getenv("RECAPTCHA_V3_SITE_KEY")
			if recaptchaV3SiteKey == "" || recaptchaV3SiteKey == "${RECAPTCHA_V3_SITE_KEY}" {
				recaptchaV3SiteKey = "6LeIxAcTAAAAAJcZVRqyHh71UMIEGNQ_MXjiZKhI" // Fallback to test key
			}

			challengeContent = `
				<div class="recaptcha-v3-wrapper" style="text-align: center; padding: 40px 20px;">
					<div class="loading-spinner" style="display: inline-block; width: 60px; height: 60px; border: 4px solid #f3f3f3; border-top: 4px solid #667eea; border-radius: 50%; animation: spin 1s linear infinite;"></div>
					<p style="margin-top: 20px; color: #4a5568; font-size: 14px;">Checking browser security...</p>
					<p style="margin-top: 10px; color: #718096; font-size: 12px;">This will only take a moment</p>
				</div>
				<script src="https://www.google.com/recaptcha/api.js?render=` + recaptchaV3SiteKey + `"></script>
				<script>
					// Auto-execute reCAPTCHA v3 on page load
					grecaptcha.ready(function() {
						grecaptcha.execute('` + recaptchaV3SiteKey + `', {action: 'bot_challenge'})
							.then(function(token) {
								console.log('reCAPTCHA v3 token obtained');
								// Set cookie to mark bot check as passed
								document.cookie = "bot_check_passed_` + domain + `=true; path=/; max-age=3600; SameSite=Lax";
								
								// Show success message
								const wrapper = document.querySelector('.recaptcha-v3-wrapper');
								wrapper.innerHTML = '<div style="color: #48bb78; font-size: 48px; margin-bottom: 15px;">✓</div><p style="color: #2d3748; font-size: 18px; font-weight: 600;">Security Check Passed</p><p style="color: #718096; font-size: 14px; margin-top: 10px;">Redirecting to application...</p>';
								
								// Redirect after short delay
								setTimeout(() => {
									window.location.reload();
								}, 1500);
							})
							.catch(function(error) {
								console.error('reCAPTCHA v3 error:', error);
								const wrapper = document.querySelector('.recaptcha-v3-wrapper');
								wrapper.innerHTML = '<div style="color: #e53e3e; padding: 15px; background: #fff5f5; border: 1px solid #feb2b2; border-radius: 8px;">⚠️ Verification failed. Please refresh the page.</div>';
							});
					});
				</script>
				<style>
					@keyframes spin {
						0% { transform: rotate(0deg); }
						100% { transform: rotate(360deg); }
					}
				</style>
			`
		} else {
			// reCAPTCHA v2 - Checkbox, visible
			challengeTitle = "reCAPTCHA Verification"
			challengeSubtitle = "Verify you're not a robot"

			// Get reCAPTCHA v2 site key from environment
			recaptchaSiteKey := os.Getenv("RECAPTCHA_SITE_KEY")
			if recaptchaSiteKey == "" || recaptchaSiteKey == "${RECAPTCHA_SITE_KEY}" {
				recaptchaSiteKey = "6LeIxAcTAAAAAJcZVRqyHh71UMIEGNQ_MXjiZKhI" // Fallback to Google's test key
			}

			challengeContent = `
				<div class="recaptcha-wrapper" style="display: flex; justify-content: center; align-items: center; min-height: 100px; margin: 30px 0;">
					<div class="g-recaptcha" data-sitekey="` + recaptchaSiteKey + `" data-callback="captchaCallback"></div>
				</div>
				<script src="https://www.google.com/recaptcha/api.js" async defer></script>
				<script>
					function captchaCallback(token) {
						if (token) {
							console.log('reCAPTCHA v2 verification successful');
							document.cookie = "bot_check_passed_` + domain + `=true; path=/; max-age=3600; SameSite=Lax";
							
							// Show success message
							const container = document.querySelector('.container');
							const successMsg = document.createElement('div');
							successMsg.style.cssText = 'margin-top: 20px; padding: 15px; background: #d4edda; border: 1px solid #c3e6cb; border-radius: 8px; color: #155724; animation: slideIn 0.3s ease-out;';
							successMsg.innerHTML = '<strong>✅ Verification successful!</strong><br>Redirecting to application...';
							container.appendChild(successMsg);
							
							setTimeout(() => {
								window.location.reload();
							}, 1500);
						}
					}
				</script>
				<noscript>
					<div style="color: #e53e3e; margin-top: 15px;">
						Please enable JavaScript to complete the verification.
					</div>
				</noscript>
			`
		}
	case "slide_puzzle":
		challengeTitle = "Puzzle Challenge"
		challengeSubtitle = "Slide the blue piece to match the dark piece"
		challengeContent = `
			<div class="puzzle-container">
				<canvas id="puzzleCanvas" width="300" height="300"></canvas>
				<input type="range" id="puzzleSlider" min="0" max="250" value="0" class="slider">
				<p class="instruction">🧩 Drag the slider to align the puzzle pieces</p>
			</div>
			<script>
				const canvas = document.getElementById('puzzleCanvas');
				const ctx = canvas.getContext('2d');
				const slider = document.getElementById('puzzleSlider');
				
				// Generate random correct position
				let correctPosition = Math.floor(Math.random() * 200) + 25;
				let attempts = 0;
				
				function drawPuzzle(offset) {
					ctx.clearRect(0, 0, 300, 300);
					
					// Background
					ctx.fillStyle = '#f7fafc';
					ctx.fillRect(0, 0, 300, 300);
					
					// Grid pattern
					ctx.strokeStyle = '#e2e8f0';
					ctx.lineWidth = 1;
					for (let i = 0; i < 300; i += 30) {
						ctx.beginPath();
						ctx.moveTo(i, 0);
						ctx.lineTo(i, 300);
						ctx.stroke();
						ctx.beginPath();
						ctx.moveTo(0, i);
						ctx.lineTo(300, i);
						ctx.stroke();
					}
					
					// Moving piece (blue)
					ctx.fillStyle = '#4299e1';
					ctx.shadowColor = 'rgba(66, 153, 225, 0.5)';
					ctx.shadowBlur = 10;
					ctx.fillRect(offset, 125, 50, 50);
					ctx.shadowBlur = 0;
					
					// Target slot (dark)
					ctx.fillStyle = '#2d3748';
					ctx.fillRect(correctPosition, 125, 50, 50);
					
					// Add puzzle piece notches
					ctx.fillStyle = '#fff';
					ctx.fillRect(offset + 20, 125 - 5, 10, 10);
					ctx.fillRect(correctPosition + 20, 125 - 5, 10, 10);
				}
				
				drawPuzzle(0);
				
				slider.addEventListener('input', function(e) {
					drawPuzzle(parseInt(e.target.value));
				});
				
				slider.addEventListener('change', function(e) {
					const offset = parseInt(e.target.value);
					attempts++;
					
					if (Math.abs(offset - correctPosition) < 5) {
						// Success!
						ctx.fillStyle = '#48bb78';
						ctx.fillRect(offset, 125, 50, 50);
						
						document.cookie = "bot_check_passed_` + domain + `=true; path=/; max-age=3600; SameSite=Lax";
						
						const instruction = document.querySelector('.instruction');
						instruction.textContent = '✅ Success! Redirecting...';
						instruction.style.color = '#48bb78';
						
						setTimeout(() => window.location.reload(), 1000);
					} else {
						// Failed
						const instruction = document.querySelector('.instruction');
						instruction.textContent = '❌ Not quite! Try again...';
						instruction.style.color = '#f56565';
						
						setTimeout(() => {
							slider.value = 0;
							drawPuzzle(0);
							instruction.textContent = '🧩 Drag the slider to align the puzzle pieces';
							instruction.style.color = '#4a5568';
						}, 1000);
					}
				});
			</script>
		`
	default:
		challengeTitle = "Verification Required"
		challengeSubtitle = "Unknown challenge type configured"
		challengeContent = `
			<div style="padding: 30px; background: #fff5f5; border: 2px solid #feb2b2; border-radius: 10px; color: #c53030;">
				<p style="margin: 0; font-weight: 600;">⚠️ Configuration Error</p>
				<p style="margin: 10px 0 0 0; font-size: 14px;">
					Bot detection is enabled but the challenge type is not properly configured.
					<br>Please contact the site administrator.
				</p>
			</div>
		`
	}

	return `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Bot Detection - DoCode WAF</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        
        .container {
            background: white;
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            max-width: 500px;
            width: 100%;
            padding: 40px;
            text-align: center;
        }
        
        .icon {
            width: 80px;
            height: 80px;
            margin: 0 auto 20px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        
        .icon svg {
            width: 40px;
            height: 40px;
            fill: white;
        }
        
        h1 {
            color: #2d3748;
            font-size: 28px;
            margin-bottom: 10px;
        }
        
        .subtitle {
            color: #718096;
            margin-bottom: 30px;
        }
        
        .puzzle-container {
            margin: 20px 0;
        }
        
        #puzzleCanvas {
            border: 2px solid #e2e8f0;
            border-radius: 10px;
            margin-bottom: 15px;
        }
        
        .slider {
            width: 100%;
            height: 8px;
            border-radius: 5px;
            background: #e2e8f0;
            outline: none;
            margin-bottom: 10px;
        }
        
        .slider::-webkit-slider-thumb {
            appearance: none;
            width: 25px;
            height: 25px;
            border-radius: 50%;
            background: #667eea;
            cursor: pointer;
        }
        
        .instruction {
            color: #4a5568;
            font-size: 14px;
        }
        
        .cf-turnstile {
            margin: 20px auto;
        }
        
        .g-recaptcha {
            display: inline-block;
            margin: 20px auto;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">
            <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
            </svg>
        </div>
        
        <h1>` + challengeTitle + `</h1>
        <p class="subtitle">` + challengeSubtitle + `</p>
        
        ` + challengeContent + `
        
        <p style="margin-top: 30px; color: #a0aec0; font-size: 14px;">
            🛡️ Protected by DoCode WAF
        </p>
    </div>
</body>
</html>`
}
