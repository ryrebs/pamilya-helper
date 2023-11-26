/// <reference types="cypress" />

context("Authentication", () => {
  beforeEach(() => {
    cy.visit("http://localhost:5000/signin");
  });

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

context("Signup", () => {
  beforeEach(() => {
    cy.visit("http://localhost:5000/signin");
  });

  after(() => {
    // Remove user
    cy.request({
      url: "/unprotected/user",
      method: "DELETE",
      body: {
        email: "aubrey@pmh.com",
      },
    }).should((response) => {
      expect(response.status).to.eq(204);
    });
  });

  it("User can signup", () => {
    // Sign up
    cy.get("a.toggle").contains("Sign up").click();
    cy.get("form[method='post'][action='/signup']").should("exist");
    cy.get("form[method='post'][action='/signup'] input[name='name']")
      .type("aubrey", { delay: 50 })
      .should("have.value", "aubrey");
    cy.get("form[method='post'][action='/signup'] input[name='email']")
      .type("aubrey@pmh.com", { delay: 50 })
      .should("have.value", "aubrey@pmh.com");
    cy.get("form[method='post'][action='/signup'] input[name='password']")
      .type("aubrey1234", { delay: 50 })
      .should("have.value", "aubrey1234");
    cy.get("form[method='post'][action='/signup']").submit();

    // Succesful signup navigates to / and can navigate to profile
    cy.location("pathname").should("equal", "/");
    cy.get("#navbarCollapse > a.btn-primary").contains("Profile").click();
    cy.location("pathname").should("equal", "/users/profile");
  });

  it("Duplicate email registration should return an error.", () => {
    // Sign up
    cy.get("a.toggle").contains("Sign up").click();
    cy.get("form[method='post'][action='/signup']").should("exist");
    cy.get("form[method='post'][action='/signup'] input[name='name']")
      .type("aubrey", { delay: 50 })
      .should("have.value", "aubrey");
    cy.get("form[method='post'][action='/signup'] input[name='email']")
      .type("aubrey@pmh.com", { delay: 50 })
      .should("have.value", "aubrey@pmh.com");
    cy.get("form[method='post'][action='/signup'] input[name='password']")
      .type("aubrey1234", { delay: 50 })
      .should("have.value", "aubrey1234");
    cy.get("form[method='post'][action='/signup']").submit();
    cy.get("div.sign-up__msg--center > span").should(
      "contain.text",
      "Invalid user e-mail. Please try again later."
    );
  });
});
