package pages

const ErrorCSS = `main {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100vh;
    width: 100vw;
}

.error-code {
    font-size: 1rem;
    color: var(--thirdary-font-color);
    font-weight: 400;
}

h1 {
    font-family: 'Instrument Serif', serif;
    font-size: 12rem;
    color: var(--primary-accent-color);
    margin-bottom: 1rem;
    font-weight: 400;
}

h2 {
    font-size: 3rem;
    color: var(--primary-font-color);
    margin-bottom: 1rem;
    font-weight: 400;
    font-style: italic;
}

.error-message {
    max-width: 600px;
    text-align: center;
    color: var(--secondary-font-color);
    margin-bottom: 2rem;
    font-weight: 400;
}

.count {
    font-size: 1rem;
    color: var(--secondary-accent-color);
    font-weight: 700;
}

button {
    background-color: black;
    color: white;
    border: none;
    padding: 0.5rem 1rem;
    border-radius: var(--border-radius-med);
    cursor: pointer;
    font-size: 1rem;
    transition: background-color 0.3s ease, color 0.3s ease;
}`
