// ***********************************************
// This example commands.js shows you how to
// create various custom commands and overwrite
// existing commands.
//
// For more comprehensive examples of custom
// commands please read more here:
// https://on.cypress.io/custom-commands
// ***********************************************
//
//
// -- This is a parent command --
// Cypress.Commands.add('login', (email, password) => { ... })
//
//
// -- This is a child command --
// Cypress.Commands.add('drag', { prevSubject: 'element'}, (subject, options) => { ... })
//
//
// -- This is a dual command --
// Cypress.Commands.add('dismiss', { prevSubject: 'optional'}, (subject, options) => { ... })
//
//
// -- This will overwrite an existing command --
// Cypress.Commands.overwrite('visit', (originalFn, url, options) => { ... })

Cypress.Commands.add("login", (email, password) => {
  cy.visit("http://localhost:5000/signin");
  cy.get("form[method='post'][action='/signin']").should("exist");
  cy.get("form[method='post'][action='/signin'] input[name='email']")
    .type(email, { delay: 50 })
    .should("have.value", email);
  cy.get("form[method='post'][action='/signin'] input[name='password']")
    .type(password, { delay: 50 })
    .should("have.value", password);
  cy.get("form[method='post'][action='/signin']").submit();
});

Cypress.Commands.add("logout", () => {
  cy.get("#user-logout-link").click();
  cy.location("pathname").should("equal", "/");
  cy.get("#navbarCollapse > a.btn-primary").should(
    "not.contain.text",
    "Profile"
  );
});

Cypress.Commands.add("register", (name, email, password) => {
  cy.visit("http://localhost:5000/signin");
  cy.get("a.toggle").contains("Sign up").click();
  cy.get("form[method='post'][action='/signup']").should("exist");
  cy.get("form[method='post'][action='/signup'] input[name='name']")
    .type(name, { delay: 50 })
    .should("have.value", name);
  cy.get("form[method='post'][action='/signup'] input[name='email']")
    .type(email, { delay: 50 })
    .should("have.value", email);
  cy.get("form[method='post'][action='/signup'] input[name='password']")
    .type(password, { delay: 50 })
    .should("have.value", password);
  cy.get("form[method='post'][action='/signup']").submit();
});

Cypress.Commands.add("removeUser", (email) => {
  cy.request({
    url: "/unprotected/user",
    method: "DELETE",
    body: {
      email,
    },
  }).should((response) => {
    expect(response.status).to.eq(204);
  });
});
