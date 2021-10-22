module.exports = {
    "roots": [
        "<rootDir>/src"
    ],
    setupFiles: [
        "<rootDir>/src/dotenv-config.ts"
    ],
    "testMatch": [
        "**/__tests__/**/*.+(ts|tsx|js)",
        "**/?(*.)+(spec|test).+(ts|tsx|js)"
    ],
    "transform": {
        "^.+\\.(ts|tsx)$": "ts-jest"
    },
}