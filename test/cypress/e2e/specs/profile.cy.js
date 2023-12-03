/// <reference types="cypress" />

context("Profile", () => {
  const email = "tester@pmh.com";
  const pwd = "tester1234";
  const name = "tester";

  before(() => {
    cy.register(name, email, pwd);
  });

  after(() => {
    cy.removeUser("tester@pmh.com");
  });

  it("Profile details is available on profile page", () => {
    cy.get("#navbarCollapse > a.btn-primary").contains("Profile").click();
    cy.location("pathname").should("equal", "/users/profile");
    cy.contains("span", name);
    cy.contains("span", email);
    cy.contains("span", "Unverified");
    cy.contains("button", "Account Verification");
  });

  it("User can request for account verification", () => {
    cy.login(email, pwd);
    cy.location("pathname").should("equal", "/");
    cy.get("#navbarCollapse > a.btn-primary").contains("Profile").click();
    cy.contains("button", "Account Verification").click();

    // Check input fields for account verification
    cy.get("input[name='name']").should("contain.value", name);
    cy.get("input[name='birthdate'][type='date']").should("exist");
    cy.get("input[name='address'][type='text']").should("exist");
    cy.get("input[name='file'][type='file']").should("exist");
    cy.get("button[type='submit'] > span").should("contain.text", "Submit");

    // Update fields for verification
    cy.fixture("img/profile.png", "binary").as("profile");
    cy.get(
      "form[method='post'][action='/users/profile/verify'][enctype='multipart/form-data']"
    ).should("exist");
    cy.get("input[name='birthdate'][type='date']").invoke(
      "attr",
      "value",
      "2020-01-01"
    );
    cy.get("input[name='address'][type='text']")
      .type("Andromeda Galaxy")
      .should("have.value", "Andromeda Galaxy");
    cy.get("input[name='file'][type='file']").selectFile("@profile");
    cy.get(
      "form[method='post'][action='/users/profile/verify'][enctype='multipart/form-data']"
    ).submit();
    cy.get("span").should("contain.text", "Success. Waiting for approval");
  });
});

context("Admin Profile", () => {
  // Verify pending account
  // Accept and Reject accounts
  // Check for verified user attributes
  // Check for rejected user attributes
});
