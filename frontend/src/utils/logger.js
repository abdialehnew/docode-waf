// Centralized logging utility to avoid console.log in production
// SonarQube recommends using a proper logging mechanism

const isDevelopment = import.meta.env.MODE === 'development'

export const logger = {
  log: (...args) => {
    if (isDevelopment) {
      console.log(...args)
    }
  },
  
  error: (...args) => {
    if (isDevelopment) {
      console.error(...args)
    }
    // In production, you might want to send errors to a logging service
    // Example: sendToErrorTracking(args)
  },
  
  warn: (...args) => {
    if (isDevelopment) {
      console.warn(...args)
    }
  },
  
  info: (...args) => {
    if (isDevelopment) {
      console.info(...args)
    }
  }
}

export default logger
