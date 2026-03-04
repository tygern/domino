import { Container } from "@cloudflare/containers";

export class DominoContainer extends Container<Env> {
  defaultPort = 8080;
  sleepAfter = "5m";
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const id = env.DOMINO_CONTAINER.idFromName("domino");
    const container = env.DOMINO_CONTAINER.get(id);
    return await container.fetch(request);
  },
};
