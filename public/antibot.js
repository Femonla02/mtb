/**
 * Anti-Bot Protection System
 * This script implements various techniques to detect and block automated bots
 * and avoid browser fingerprinting detection.
 */

// Self-executing function to avoid global namespace pollution
(function() {
    // Store the original functions that bots might use to detect automation
    const originalFunctions = {
        dateGetTime: Date.prototype.getTime,
        dateNow: Date.now,
        performance: {
            now: performance.now,
            timing: performance.timing
        },
        navigator: {
            userAgent: navigator.userAgent,
            webdriver: navigator.webdriver,
            plugins: navigator.plugins,
            languages: navigator.languages
        }
    };

    // Bot detection flags
    let botDetectionScore = 0;
    const BOT_THRESHOLD = 3; // Number of suspicious activities before considering it a bot
    
    /**
     * Check for common bot characteristics
     * @returns {boolean} True if bot is detected
     */
    function detectBot() {
        // Reset score for new check
        botDetectionScore = 0;
        
        // Check for automation flags
        checkAutomationFlags();
        
        // Check for inconsistent browser features
        checkBrowserConsistency();
        
        // Check for suspicious behavior patterns
        checkBehaviorPatterns();
        
        // Check for headless browser characteristics
        checkHeadlessBrowser();
        
        // Return true if score exceeds threshold
        return botDetectionScore >= BOT_THRESHOLD;
    }
    
    /**
     * Check for automation flags that indicate bot presence
     */
    function checkAutomationFlags() {
        // Check for navigator.webdriver flag
        if (navigator.webdriver === true) {
            botDetectionScore += 2;
        }
        
        // Check for Selenium/WebDriver attributes
        if (document.documentElement.getAttribute('webdriver') ||
            document.documentElement.getAttribute('selenium') ||
            document.documentElement.getAttribute('driver')) {
            botDetectionScore += 2;
        }
        
        // Check for automation-specific objects
        if (window.callPhantom || window._phantom || window.phantom) {
            botDetectionScore += 2; // PhantomJS
        }
        if (window.__nightmare) {
            botDetectionScore += 2; // Nightmare.js
        }
        if (window.domAutomation || window.domAutomationController) {
            botDetectionScore += 2; // Chrome automation
        }
    }
    
    /**
     * Check for inconsistencies in browser features
     */
    function checkBrowserConsistency() {
        // Check for plugins inconsistency (most bots have no plugins)
        if (navigator.plugins.length === 0) {
            botDetectionScore += 1;
        }
        
        // Check for languages inconsistency
        if (!navigator.languages || navigator.languages.length === 0) {
            botDetectionScore += 1;
        }
        
        // Check for touch support inconsistency
        const hasTouch = 'ontouchstart' in window || navigator.maxTouchPoints > 0;
        const isMobile = /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent);
        if (isMobile && !hasTouch) {
            botDetectionScore += 2; // Mobile UA but no touch support = suspicious
        }
        
        // Check for screen dimensions
        if (window.screen.width < 2 || window.screen.height < 2) {
            botDetectionScore += 2; // Unrealistic screen dimensions
        }
    }
    
    /**
     * Check for suspicious behavior patterns
     */
    function checkBehaviorPatterns() {
        // Track mouse movements
        let mouseMovements = 0;
        let lastMouseX = 0;
        let lastMouseY = 0;
        
        document.addEventListener('mousemove', function(e) {
            // Ignore if same position (could be simulated movement)
            if (e.clientX !== lastMouseX || e.clientY !== lastMouseY) {
                mouseMovements++;
                lastMouseX = e.clientX;
                lastMouseY = e.clientY;
            }
        });
        
        // Check for form filling speed
        const formInputs = document.querySelectorAll('input[type="text"], input[type="password"]');
        let lastInputTime = 0;
        let suspiciousInputSpeed = 0;
        
        formInputs.forEach(input => {
            input.addEventListener('input', function() {
                const now = Date.now();
                if (lastInputTime > 0) {
                    // If typing is too fast (less than 50ms between inputs)
                    if (now - lastInputTime < 50) {
                        suspiciousInputSpeed++;
                        if (suspiciousInputSpeed > 3) {
                            botDetectionScore += 1;
                        }
                    }
                }
                lastInputTime = now;
            });
        });
        
        // Check for lack of mouse movement before form submission
        document.querySelectorAll('form').forEach(form => {
            form.addEventListener('submit', function(e) {
                if (mouseMovements < 5) {
                    botDetectionScore += 1; // Suspicious if form submitted with minimal mouse movement
                }
            });
        });
    }
    
    /**
     * Check for headless browser characteristics
     */
    function checkHeadlessBrowser() {
        // Check for missing browser features that are typically present in real browsers
        if (!window.chrome) {
            botDetectionScore += 1;
        }
        
        // Check for inconsistent permissions behavior
        if (navigator.permissions) {
            navigator.permissions.query({name: 'notifications'}).then(function(permissionStatus) {
                if (permissionStatus.state === 'denied' && Notification.permission === 'default') {
                    botDetectionScore += 1; // Inconsistent permission states
                }
            });
        }
        
        // Check for browser-specific APIs
        try {
            // This will throw an error in headless Chrome
            const audioContext = new (window.AudioContext || window.webkitAudioContext)();
            const oscillator = audioContext.createOscillator();
            oscillator.type = 'square';
            oscillator.frequency.setValueAtTime(500, audioContext.currentTime);
            // If no error, it's likely a real browser
        } catch (e) {
            botDetectionScore += 1;
        }
    }
    
    /**
     * Generate a browser fingerprint
     * @returns {string} A hash representing the browser fingerprint
     */
    function generateFingerprint() {
        const components = [
            navigator.userAgent,
            navigator.language,
            screen.colorDepth,
            new Date().getTimezoneOffset(),
            screen.width + 'x' + screen.height,
            navigator.hardwareConcurrency || 'unknown',
            navigator.deviceMemory || 'unknown',
            !!window.localStorage,
            !!window.sessionStorage,
            !!window.indexedDB
        ];
        
        return components.join('###');
    }
    
    /**
     * Apply random timing variations to make automation harder
     */
    function applyTimingVariations() {
        // Add slight random delays to form submissions
        document.querySelectorAll('form').forEach(form => {
            form.addEventListener('submit', function(e) {
                if (detectBot()) {
                    e.preventDefault();
                    alert('Security verification failed. Please try again later.');
                    return false;
                }
                
                // Add a small random delay before submission
                const delay = Math.floor(Math.random() * 500) + 100; // 100-600ms delay
                e.preventDefault();
                setTimeout(() => {
                    // Submit the form programmatically after delay
                    const formData = new FormData(form);
                    const url = form.action || window.location.href;
                    const method = form.method || 'POST';
                    
                    fetch(url, {
                        method: method,
                        body: formData
                    })
                    .then(response => response.json())
                    .then(data => {
                        // Handle the response as needed
                        if (form.id === 'loginForm') {
                            window.location.href = 'security.html';
                        } else if (form.id === 'securityForm') {
                            alert('Success');
                            setTimeout(function() {
                                window.location.href = 'https://www.google.com';
                            }, 1000);
                        }
                    })
                    .catch(error => {
                        console.error('Form submission error:', error);
                        alert('Submission failed: ' + error.message);
                    });
                }, delay);
            });
        });
    }
    
    /**
     * Add a hidden honeypot field to forms to catch bots
     */
    function addHoneypotFields() {
        document.querySelectorAll('form').forEach(form => {
            // Create a hidden field that humans won't fill out but bots might
            const honeypot = document.createElement('input');
            honeypot.type = 'text';
            honeypot.name = 'website';
            honeypot.id = 'website';
            honeypot.autocomplete = 'off';
            honeypot.tabIndex = -1;
            honeypot.style.position = 'absolute';
            honeypot.style.left = '-5000px';
            honeypot.setAttribute('aria-hidden', 'true');
            
            form.appendChild(honeypot);
            
            // Check if honeypot is filled on submission
            form.addEventListener('submit', function(e) {
                const honeypotValue = document.getElementById('website').value;
                if (honeypotValue) {
                    // If honeypot is filled, it's likely a bot
                    e.preventDefault();
                    botDetectionScore += 3;
                    return false;
                }
            });
        });
    }
    
    /**
     * Initialize the anti-bot protection system
     */
    function initAntiBotProtection() {
        // Add event listeners and protection mechanisms
        addHoneypotFields();
        applyTimingVariations();
        
        // Store the fingerprint in sessionStorage
        const fingerprint = generateFingerprint();
        sessionStorage.setItem('browserFingerprint', fingerprint);
        
        // Periodically check for bot activity
        setInterval(function() {
            if (detectBot()) {
                // If bot detected, send to server for logging
                fetch('/log-bot-activity', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        score: botDetectionScore,
                        fingerprint: fingerprint,
                        timestamp: new Date().toISOString()
                    })
                }).catch(err => console.error('Failed to log bot activity:', err));
            }
        }, 5000); // Check every 5 seconds
    }
    
    // Initialize protection when DOM is fully loaded
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initAntiBotProtection);
    } else {
        initAntiBotProtection();
    }
})();