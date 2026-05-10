import { Pool, type PoolClient, type QueryResult, type QueryResultRow } from "pg";
import { env } from "../support/env";

let pool: Pool | null = null;

function getPool(): Pool {
  if (!pool) {
    pool = new Pool({
      connectionString: env.databaseUrl
    });
  }
  return pool;
}

export async function withDbClient<T>(fn: (client: PoolClient) => Promise<T>): Promise<T> {
  const client = await getPool().connect();
  try {
    return await fn(client);
  } finally {
    client.release();
  }
}

export async function dbQuery<T extends QueryResultRow = QueryResultRow>(
  sql: string,
  values: unknown[] = []
): Promise<QueryResult<T>> {
  return getPool().query<T>(sql, values);
}

export async function dbTransaction<T>(fn: (client: PoolClient) => Promise<T>): Promise<T> {
  return withDbClient(async (client) => {
    await client.query("BEGIN");
    try {
      const result = await fn(client);
      await client.query("COMMIT");
      return result;
    } catch (error) {
      await client.query("ROLLBACK");
      throw error;
    }
  });
}
