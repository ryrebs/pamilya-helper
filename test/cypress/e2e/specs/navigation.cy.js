/// <reference types="cypress" />

context("Navigation", () => {
  beforeEach(() => {
    cy.visit("http://localhost:5000");
  });

  it("Index page should contain the navigation", () => {
    cy.get(".navbar").contains("PamilyaHelper").click();
    cy.get("#navbarCollapse").contains("Home");
    cy.get("#navbarCollapse").contains("About");
    cy.get("#navbarCollapse").contains("Job List");
    cy.get("#navbarCollapse").contains("Helpers");
    cy.get("#navbarCollapse").contains("Contact");
    cy.get("#navbarCollapse").contains("Login/SignUp");
  });

  it("Home navigation should go to /", () => {
    cy.get("#navbarCollapse > div > a.nav-item.nav-link.active")
      .contains("Home")
      .click();
    cy.get(".navbar").contains("PamilyaHelper").click();
  });

  it("About navigation should go to /about", () => {
    cy.get("#navbarCollapse > div > a.nav-item.nav-link")
      .contains("About")
      .click();
    cy.location("pathname").should("include", "about");
  });

  it("Login/Signup should go to login/signup page.", () => {
    cy.get("#navbarCollapse > a.btn-primary").contains("Login/SignUp").click();
    cy.location("pathname").should("include", "signin");

    cy.get("form[method='post'][action='/signin']").should("exist")
    cy.get("form[method='post'][action='/signin'] input[name='email'] + label").should(
      "have.text",
      "Email"
    );
    cy.get("form[method='post'][action='/signin'] input[name='password'] + label").should(
      "have.text",
      "Password"
    );
  });

  it("cy.reload() - reload the page", () => {
    cy.reload();
    cy.reload(true); // reload the page without using the cache
  });
});
