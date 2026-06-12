package pages

const ErrorCSS = `body {
    height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    background: radial-gradient(circle at top, #1a1a2e, #0f0f1a);
    color: #fff;
}

.error-container {
    width: 90%;
    max-width: 500px;
    padding: 32px;
    text-align: center;

    background: rgba(255, 255, 255, 0.06);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 16px;

    backdrop-filter: blur(12px);
    box-shadow: 0 10px 40px rgba(0, 0, 0, 0.5);

    animation: fadeIn 0.5s ease-out;
}

h1 {
    font-size: 2.5rem;
    margin-bottom: 12px;
    letter-spacing: 1px;
}

p {
    opacity: 0.8;
    margin: 8px 0;
}

h1::before {
    content: "";
    display: block;
    width: 60px;
    height: 4px;
    margin: 0 auto 12px auto;
    border-radius: 10px;
    background: linear-gradient(90deg, #ff4d4d, #ffb84d);
}

.countdown {
    margin-top: 16px;
    font-weight: bold;
    font-size: 1.1rem;
    color: #4dd0ff;
}

button {
    margin-top: 20px;
    padding: 10px 18px;
    border: none;
    border-radius: 10px;
    cursor: pointer;

    background: #4dd0ff;
    color: #000;
    font-weight: bold;

    transition: 0.2s ease;
}

button:hover {
    transform: scale(1.05);
    background: #7ae3ff;
}

@keyframes fadeIn {
    from {
        opacity: 0;
        transform: translateY(10px);
    }

    to {
        opacity: 1;
        transform: translateY(0);
    }
}`
