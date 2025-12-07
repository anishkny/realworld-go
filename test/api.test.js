import { describe, it, before, after } from "mocha";
import { faker } from "@faker-js/faker";
import { strict as assert } from "assert";
import Ajv from "ajv";
import addFormats from "ajv-formats";
import axios from "axios";
import fs from "fs";

const BASE_URL = "http://localhost:3000/api";
const CONTEXT_FILE = "test/context.json";

axios.defaults.validateStatus = (status) => status < 500;
axios.defaults.baseURL = BASE_URL;

const ajv = new Ajv({ allErrors: true });
addFormats(ajv);

const context = {};

before(() => {
  if (fs.existsSync(CONTEXT_FILE)) {
    Object.assign(context, JSON.parse(fs.readFileSync(CONTEXT_FILE, "utf-8")));
  }

  // Create test user if not exists
  if (!context.user) {
    context.user = generateTestUserData();
  }
});

after(() => {
  fs.writeFileSync(CONTEXT_FILE, JSON.stringify(context, null, 2));
});

describe("Health", () => {
  it("Root", async () => {
    const res = await axios.get("");
    assert.equal(res.status, 200);
  });
});

describe("Auth", () => {
  it("Register", async () => {
    context.user = generateTestUserData();
    const res = await axios.post("/users", { user: context.user });
    assert.equal(res.status, 200);
    assertSchema(res.data, getSchemas().authenticatedUser);
    assert.equal(res.data.user.username, context.user.username);
    assert.equal(res.data.user.email, context.user.email);
    context.user.token = res.data.user.token;
  });

  it("Register - Bad request", async () => {
    const res = await axios.post("/users", {
      user: { email: context.user.email },
    });
    assert.equal(res.status, 422);
    assert.deepEqual(res.data, {
      errors: {
        Username:
          "Field validation for 'Username' failed on the 'required' tag",
        Password:
          "Field validation for 'Password' failed on the 'required' tag",
      },
    });
  });

  it("Register - Empty request", async () => {
    const res = await axios.post("/users", null);
    assert.equal(res.status, 422);
    assert.deepEqual(res.data, { errors: { error: "EOF" } });
  });

  it("Login", async () => {
    const res = await axios.post("/users/login", {
      user: {
        email: context.user.email,
        password: context.user.password,
      },
    });
    assert.equal(res.status, 200);
    assertSchema(res.data, getSchemas().authenticatedUser);
    assert.equal(res.data.user.username, context.user.username);
    assert.equal(res.data.user.email, context.user.email);
    context.user.token = res.data.user.token;
  });

  it("Login - Bad request", async () => {
    const res = await axios.post("/users/login", {
      user: { email: context.user.email },
    });
    assert.equal(res.status, 422);
    assert.deepEqual(res.data, {
      errors: {
        Password:
          "Field validation for 'Password' failed on the 'required' tag",
      },
    });
  });

  it("Login - Empty request", async () => {
    const res2 = await axios.post("/users/login", null);
    assert.equal(res2.status, 422);
    assert.deepEqual(res2.data, { errors: { error: "EOF" } });
  });

  it("Login - Unknown email", async () => {
    const res = await axios.post("/users/login", {
      user: {
        email: faker.internet.email(),
        password: context.user.password,
      },
    });
    assert.equal(res.status, 401);
  });

  it("Login - Wrong password", async () => {
    const res = await axios.post("/users/login", {
      user: {
        email: context.user.email,
        password: "wrongpassword",
      },
    });
    assert.equal(res.status, 401);
  });
});

describe("User", () => {
  it("Current user", async () => {
    const res = await axios.get("/user", {
      headers: { Authorization: context.user.token },
    });
    assert.equal(res.status, 200);
    assertSchema(res.data, getSchemas().authenticatedUser);
    assert.equal(res.data.user.username, context.user.username);
    assert.equal(res.data.user.email, context.user.email);
  });

  it("Current user - Bad token", async () => {
    const res = await axios.get("/user", {
      headers: { Authorization: "BadToken" },
    });
    assert.equal(res.status, 401);
  });

  it("Current user - Missing token", async () => {
    const res = await axios.get("/user");
    assert.equal(res.status, 401);
  });

  it("Update user", async () => {
    const newUsername = "updated_" + context.user.username;
    const newEmail = "updated_" + context.user.email;
    const newPassword = "newpassword";
    const newBio = faker.lorem.sentence();
    const newImage = faker.image.avatar();
    const res = await axios.put(
      "/user",
      {
        user: {
          email: newEmail,
          username: newUsername,
          password: newPassword,
          bio: newBio,
          image: newImage,
        },
      },
      { headers: { Authorization: context.user.token } }
    );
    assert.equal(res.status, 200);
    assertSchema(res.data, getSchemas().authenticatedUser);
    assert.equal(res.data.user.bio, newBio);
    assert.equal(res.data.user.image, newImage);
    context.user = res.data.user;
    context.user.password = newPassword;
  });

  it("Update user - Bad request", async () => {
    const res = await axios.put("/user", `{ gibberish `, {
      headers: { Authorization: context.user.token },
    });
    assert.equal(res.status, 422);
    assert.deepEqual(res.data, {
      errors: {
        error:
          "invalid character 'g' looking for beginning of object key string",
      },
    });
  });

  it("Update user - No mutations", async () => {
    const res = await axios.put(
      "/user",
      { user: {} },
      { headers: { Authorization: context.user.token } }
    );
    assert.equal(res.status, 422);
    assert.deepEqual(res.data, {
      error: "At least one field must be provided",
    });
  });
});

describe("Profile", () => {
  before(async () => {
    // Register celeb user
    context.celebUser = generateTestUserData("celeb_");
    const res = await axios.post("/users", { user: context.celebUser });
    assert.equal(res.status, 200);
    context.celebUser.token = res.data.user.token;
  });

  it("Get profile", async () => {
    const res = await axios.get(`/profiles/${context.celebUser.username}`, {
      headers: { Authorization: context.user.token },
    });
    assert.equal(res.status, 200);
    assertSchema(res.data, getSchemas().profile);
    assert.equal(res.data.profile.username, context.celebUser.username);
    assert.equal(res.data.profile.bio, "");
    assert.equal(res.data.profile.image, "");
    assert.equal(res.data.profile.following, false);
  });

  it("Get profile - Unauthenticated", async () => {
    const res = await axios.get(`/profiles/${context.celebUser.username}`);
    assert.equal(res.status, 200);
    assertSchema(res.data, getSchemas().profile);
    assert.equal(res.data.profile.username, context.celebUser.username);
    assert.equal(res.data.profile.bio, "");
    assert.equal(res.data.profile.image, "");
    assert.equal(res.data.profile.following, false);
  });

  it("Get profile - Unknown user", async () => {
    const res = await axios.get(`/profiles/unknownuser`, {
      headers: { Authorization: context.user.token },
    });
    assert.equal(res.status, 404);
  });

  it("Follow", async () => {
    const res = await axios.post(
      `/profiles/${context.celebUser.username}/follow`,
      {},
      { headers: { Authorization: context.user.token } }
    );
    assert.equal(res.status, 200);
    assertSchema(res.data, getSchemas().profile);
    assert.equal(res.data.profile.username, context.celebUser.username);
    assert.equal(res.data.profile.following, true);
  });

  it("Profile after follow", async () => {
    const res = await axios.get(`/profiles/${context.celebUser.username}`, {
      headers: { Authorization: context.user.token },
    });
    assert.equal(res.status, 200);
    assertSchema(res.data, getSchemas().profile);
    assert.equal(res.data.profile.username, context.celebUser.username);
    assert.equal(res.data.profile.following, true);
  });

  it("Follow unknown", async () => {
    const res = await axios.post(
      `/profiles/${faker.internet.username()}/follow`,
      {},
      { headers: { Authorization: context.user.token } }
    );
    assert.equal(res.status, 404);
  });

  it("Unfollow", async () => {
    const res = await axios.delete(
      `/profiles/${context.celebUser.username}/follow`,
      { headers: { Authorization: context.user.token } }
    );
    assert.equal(res.status, 200);
    assertSchema(res.data, getSchemas().profile);
    assert.equal(res.data.profile.username, context.celebUser.username);
    assert.equal(res.data.profile.following, false);
  });

  it("Profile after unfollow", async () => {
    const res = await axios.get(`/profiles/${context.celebUser.username}`, {
      headers: { Authorization: context.user.token },
    });
    assert.equal(res.status, 200);
    assertSchema(res.data, getSchemas().profile);
    assert.equal(res.data.profile.username, context.celebUser.username);
    assert.equal(res.data.profile.following, false);
  });

  it("Unfollow unknown", async () => {
    const res = await axios.delete(
      `/profiles/${faker.internet.username()}/follow`,
      { headers: { Authorization: context.user.token } }
    );
    assert.equal(res.status, 404);
  });
});

// ----------------------------------------
// HELPERS
// ----------------------------------------
function generateTestUserData(prefix = "") {
  const username = prefix + faker.internet.username().toLowerCase();
  const email = `${username}@email.com`;
  const password = "password";
  return { username, email, password };
}

function assertSchema(data, schema) {
  const validate = ajv.compile(schema);
  const valid = validate(data);
  if (!valid) {
    console.error(validate.errors);
  }
  assert.ok(valid, "Response does not match schema");
}

function getSchemas() {
  return {
    authenticatedUser: {
      type: "object",
      properties: {
        user: {
          type: "object",
          properties: {
            email: {
              type: "string",
            },
            token: {
              type: "string",
            },
            username: {
              type: "string",
            },
            bio: {
              type: ["string", "null"],
            },
            image: {
              type: ["string", "null"],
            },
          },
          additionalProperties: false,
          required: ["email", "token", "username"],
        },
      },
      additionalProperties: false,
      required: ["user"],
    },
    profile: {
      type: "object",
      properties: {
        profile: {
          type: "object",
          properties: {
            username: {
              type: "string",
            },
            bio: {
              type: "string",
            },
            image: {
              type: "string",
            },
            following: {
              type: "boolean",
            },
          },
          required: ["username", "bio", "image", "following"],
        },
      },
      required: ["profile"],
    },
    article: {
      type: "object",
      properties: {
        article: {
          type: "object",
          properties: {
            slug: { type: "string" },
            title: { type: "string" },
            description: { type: "string" },
            body: { type: "string" },
            tagList: {
              type: "array",
              items: { type: "string" },
            },
            createdAt: {
              type: "string",
              format: "date-time",
            },
            updatedAt: {
              type: "string",
              format: "date-time",
            },
            favorited: { type: "boolean" },
            favoritesCount: { type: "integer" },
            author: {
              type: "object",
              properties: {
                username: { type: "string" },
                bio: { type: "string" },
                image: { type: "string" },
                following: { type: "boolean" },
              },
              required: ["username", "bio", "image", "following"],
              additionalProperties: false,
            },
          },
          required: [
            "slug",
            "title",
            "description",
            "body",
            "tagList",
            "createdAt",
            "updatedAt",
            "favorited",
            "favoritesCount",
            "author",
          ],
          additionalProperties: false,
        },
      },
      required: ["article"],
      additionalProperties: false,
    },
  };
}

function assertArticlesInDescendingOrder(articles) {
  for (let i = 1; i < articles.length; i++) {
    const prevDate = new Date(articles[i - 1].updatedAt);
    const currDate = new Date(articles[i].updatedAt);
    assert.ok(prevDate >= currDate);
  }
}
