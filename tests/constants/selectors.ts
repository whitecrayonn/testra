export const SELECTORS = {
  emailInput: { label: "Email" },
  passwordInput: { label: "Password" },
  nameInput: { label: "Full name" },
  loginButton: { name: "Sign in" },
  registerButton: { name: "Create account" },
  submitButton: { name: /create|save|submit|sign in/i },
  alert: "[role=alert]",
  pageHeader: (title: string) => `h1:has-text("${title}"), [role=heading]:has-text("${title}")`,
  card: "[class*=Card]",
  errorCard: "text=Failed to",
};
