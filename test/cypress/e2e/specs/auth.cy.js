/// <reference types="cypress" />

context("Authentication", () => {
  beforeEach(() => {
    cy.visit("http://localhost:5000/signin");
  });

  // signin should go to index

  it("Existing can user can signin and signout", () => {
    // Sign in
    cy.get("form[method='post'][action='/signin']").should("exist");
    cy.get("form[method='post'][action='/signin'] input[name='email']")
      .type("admin@admin.com", { delay: 50 })
      .should("have.value", "admin@admin.com");
    cy.get("form[method='post'][action='/signin'] input[name='password']")
      .type("admin1234", { delay: 50 })
      .should("have.value", "admin1234");
    cy.get("form[method='post'][action='/signin']").submit();

    // navigate to profile
    cy.location("pathname").should("equal", "/");
    cy.get("#navbarCollapse > a.btn-primary").contains("Profile").click();
    cy.location("pathname").should("equal", "/users/profile");

    // navigate to /signin
    cy.visit("http://localhost:5000/signin")
      .location("pathname")
      .should("equal", "/users/profile");

    // signout
    cy.get("#user-logout-link").click();
    cy.location("pathname").should("equal", "/");
    cy.get("#navbarCollapse > a.btn-primary").should(
      "not.contain.text",
      "Profile"
    );
  });

  it("Non Existing  user cannot signin", () => {
    // Sign in
    cy.get("form[method='post'][action='/signin'] input[name='email']")
      .type("notexisting@pm.com", { delay: 50 })
      .should("have.value", "notexisting@pm.com");
    cy.get("form[method='post'][action='/signin'] input[name='password']")
      .type("notexisting", { delay: 50 })
      .should("have.value", "notexisting");
    cy.get("form[method='post'][action='/signin']").submit();
    cy.get("div.sign-in__msg--center > span").should(
      "contain.text",
      "Invalid Email or Password"
    );

    // Visit profile page
    cy.visit("http://localhost:5000/users/profile")
      .location("pathname")
      .should("equal", "/signin");
  });
});
