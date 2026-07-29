export const TEST_DATA = {
  users: {
    valid: {
      name: "Test User",
      email: "test.user@example.com",
      password: "SecurePass123!",
    },
    invalid: {
      shortPassword: "123",
      noUppercase: "lowercase123!",
      noNumber: "NoNumberHere!",
      noSpecial: "NoSpecialChar123",
    },
  },
  projects: {
    valid: {
      name: "E-commerce Platform",
      key: "ECOM",
      description: "Main e-commerce testing project",
    },
    invalid: {
      emptyName: "",
      emptyKey: "",
    },
  },
  testCases: {
    valid: {
      title: "Verify user can add item to cart",
      description: "As a user, I should be able to add items to my shopping cart",
      status: "draft",
      priority: "medium",
      tags: ["smoke", "cart"],
      steps: [
        { action: "Navigate to product page", expected: "Product page loads", test_data: "" },
        { action: "Click Add to Cart", expected: "Item added to cart", test_data: "" },
      ],
    },
    invalid: {
      emptyTitle: "",
    },
  },
  defects: {
    valid: {
      title: "Login button not responding on mobile",
      description: "The login button on mobile devices does not respond to tap events",
      severity: "high",
      priority: "high",
    },
  },
  apiCollections: {
    valid: {
      name: "User API Tests",
      description: "Collection of API tests for user endpoints",
    },
    requests: {
      valid: {
        name: "Get User Profile",
        method: "GET",
        url: "https://httpbin.org/get",
        headers: { "Accept": "application/json" },
      },
      invalid: {
        name: "Invalid Request",
        method: "GET",
        url: "not-a-valid-url",
      },
    },
  },
  automation: {
    valid: {
      name: "Web Automation Suite",
      framework: "playwright",
      repositoryUrl: "https://github.com/testra/web-automation",
      branch: "main",
      command: "npx playwright test",
    },
  },
  integrations: {
    valid: {
      name: "Slack Notifications",
      provider: "slack",
      config: { webhook_url: "https://hooks.slack.com/services/FAKE" },
    },
  },
};
