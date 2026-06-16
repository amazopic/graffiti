import { Session } from './auth/session.js';

export class App {
  run(req) {
    return validate(req);
  }
}

function validate(req) {
  return Boolean(req);
}
