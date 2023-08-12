
document.getElementById("forgotPasswordForm").addEventListener("submit", async function(event) {
    event.preventDefault();
    
    let traceId = generateUUID();
    const spanId = "forgot-password-page";

    const email = document.getElementById("email").value;

    console.log("Email:", email, "\nTrace ID:", traceId, "\nSpan ID:", spanId);

    // Send a request to reset password using fetch or any other library
    try {
            const response = await fetch("http://localhost:3000/users/reset-password", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                    "X-Trace-Id": traceId,
                    "X-Parent-Id": spanId
                },
                body: JSON.stringify({ email })
            });

            const data = await response.json();

            // Display success or error message
            const messageElement = document.getElementById("message");
            if (response.ok) {
                messageElement.style.color = "green";
                messageElement.textContent = data.message;
            } else {
                messageElement.style.color = "red";
                messageElement.textContent = data.error;
            }
    } catch (error) {
        console.error("Error:", error);
    }
});

function generateUUID() {
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
        var r = Math.random() * 16 | 0,
            v = c === 'x' ? r : (r & 0x3 | 0x8);
        return v.toString(16);
    });
}