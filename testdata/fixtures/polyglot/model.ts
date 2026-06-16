import { Session } from './auth/session';

export interface User { id: number }

export class Model {
  load(id: number): User {
    return fetchUser(id);
  }
}

function fetchUser(id: number): User {
  return { id };
}
