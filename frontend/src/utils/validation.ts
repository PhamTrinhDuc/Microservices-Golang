export const validateEmail = (email: string): string | null => {
  const emailRegex = /^(?!.*\.\.)[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$/
  if (!email) {
    return 'Email is required'
  }
  if (!emailRegex.test(email)) {
    return 'Invalid email format'
  }
  return null
}

export const validatePassword = (password: string): string | null => {
  if (!password) {
    return 'Password is required'
  }
  if (password.length < 8) {
    return 'Password must be at least 8 characters'
  }
  return null
}

export const validateName = (name: string): string | null => {
  if (!name.trim()) {
    return 'Name is required'
  }
  if (name.trim().length < 2) {
    return 'Name must be at least 2 characters'
  }
  return null
}

export const getPasswordStrength = (password: string): {
  score: number
  label: 'Weak' | 'Medium' | 'Strong'
} => {
  let score = 0
  if (password.length >= 8) score += 1
  if (/[A-Z]/.test(password)) score += 1
  if (/[0-9]/.test(password)) score += 1
  if (/[^A-Za-z0-9]/.test(password)) score += 1

  if (score <= 1) {
    return { score, label: 'Weak' }
  }

  if (score <= 3) {
    return { score, label: 'Medium' }
  }

  return { score, label: 'Strong' }
}
