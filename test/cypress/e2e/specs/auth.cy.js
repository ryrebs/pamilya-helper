/// <reference types="cypress" />

context("Authentication", () => {
  beforeEach(() => {
    cy.visit("http://localhost:5000/signin");
  });

  it("Existing can user can signin and signout", () => {
    // Sign in
    cy.login("admin@pmh.com", "admin1234");
    cy.location("pathname").should("equal", "/");

    // navigate to profile
    cy.get("#navbarCollapse > a.btn-primary").contains("Profile").click();
    cy.location("pathname").should("equal", "/users/profile");

    // navigate to /signin
    cy.visit("http://localhost:5000/signin")
      .location("pathname")
      .should("equal", "/users/profile");

    // signout
    cy.logout();
  });

  it("Non Existing  user cannot signin", () => {
    // Sign in
    cy.login("notexisting@pm.com", "notexisting");
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

context("Signup", () => {
  beforeEach(() => {
    cy.visit("http://localhost:5000/signin");
  });

  after(() => {
    // Remove user
    cy.removeUser("tester@pmh.com");
  });

  it("User can signup", () => {
    // Sign up
    cy.register("tester", "tester@pmh.com", "tester1234");

    // Succesful signup navigates to / and can navigate to profile
    cy.location("pathname").should("equal", "/");
    cy.get("#navbarCollapse > a.btn-primary").contains("Profile").click();
    cy.location("pathname").should("equal", "/users/profile");
  });

  it("Duplicate email registration should return an error.", () => {
    // Sign up
    cy.register("tester", "tester@pmh.com", "tester1234");
    cy.get("div.sign-up__msg--center > span").should(
      "contain.text",
      "Invalid user e-mail. Please try again later."
    );
  });
});
