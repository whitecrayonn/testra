import { ApiHelper } from "../helpers/api";
import { uniqueEmail, generatePassword, uniqueId } from "../helpers/random";
import { User } from "../types";

export interface CreatedUser {
  user: User;
  password: string;
}

export class UserFactory {
  constructor(private api: ApiHelper) {}

  build(input: Partial<CreatedUser> = {}): { email: string; password: string; name: string } {
    return {
      email: input.user?.email || uniqueEmail(),
      password: input.password || generatePassword(),
      name: input.user?.name || `Test User ${uniqueId()}`,
    };
  }

  async create(input: Partial<CreatedUser> = {}): Promise<CreatedUser> {
    const body = this.build(input);
    const result = await this.api.register(body);
    return { user: result.user, password: body.password };
  }
}
