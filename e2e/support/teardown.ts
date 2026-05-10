import { dbTransaction } from "../fixtures/db";

export type CreatedEntities = {
  memberships: string[];
  positions: string[];
  divisions: string[];
  roles: string[];
  users: string[];
};

export function createCreatedEntities(): CreatedEntities {
  return {
    memberships: [],
    positions: [],
    divisions: [],
    roles: [],
    users: []
  };
}

export function trackCreated(entities: CreatedEntities, kind: keyof CreatedEntities, id: string): void {
  if (!id) {
    return;
  }
  if (!entities[kind].includes(id)) {
    entities[kind].push(id);
  }
}

function membershipKeyToColumns(key: string): { userId: string; positionId: string } {
  const [userId, positionId] = key.split("::");
  return { userId, positionId };
}

export async function teardownCreatedEntities(entities: CreatedEntities): Promise<void> {
  await dbTransaction(async (client) => {
    const membershipSubjectIds = entities.memberships
      .map((membershipKey) => {
        const { userId, positionId } = membershipKeyToColumns(membershipKey);
        return userId && positionId ? `${userId}:${positionId}` : "";
      })
      .filter(Boolean);

    await client.query("ALTER TABLE event_log DISABLE TRIGGER event_log_no_delete");
    try {
      if (membershipSubjectIds.length > 0) {
        await client.query("DELETE FROM event_log WHERE subject_id = ANY($1::text[])", [membershipSubjectIds]);
      }

      if (entities.positions.length > 0) {
        await client.query("DELETE FROM event_log WHERE subject_id = ANY($1::text[])", [entities.positions]);
      }

      if (entities.divisions.length > 0) {
        await client.query("DELETE FROM event_log WHERE subject_id = ANY($1::text[])", [entities.divisions]);
      }

      if (entities.roles.length > 0) {
        await client.query("DELETE FROM event_log WHERE subject_id = ANY($1::text[])", [entities.roles]);
      }

      if (entities.users.length > 0) {
        await client.query(
          "DELETE FROM event_log WHERE actor_id = ANY($1::text[]) OR subject_id = ANY($1::text[]) OR split_part(subject_id, ':', 1) = ANY($1::text[])",
          [entities.users]
        );
      }
    } finally {
      await client.query("ALTER TABLE event_log ENABLE TRIGGER event_log_no_delete");
    }

    if (entities.memberships.length > 0) {
      for (const membershipKey of entities.memberships) {
        const { userId, positionId } = membershipKeyToColumns(membershipKey);
        if (!userId || !positionId) {
          continue;
        }
        await client.query("DELETE FROM memberships WHERE user_id = $1 AND position_id = $2", [userId, positionId]);
      }
    }

    if (entities.positions.length > 0) {
      await client.query("DELETE FROM positions WHERE id = ANY($1::text[])", [entities.positions]);
    }

    if (entities.divisions.length > 0) {
      await client.query("DELETE FROM divisions WHERE id = ANY($1::text[])", [entities.divisions]);
    }

    if (entities.roles.length > 0) {
      await client.query("DELETE FROM roles WHERE id = ANY($1::text[])", [entities.roles]);
    }

    if (entities.users.length > 0) {
      await client.query("DELETE FROM users WHERE id = ANY($1::text[])", [entities.users]);
    }
  });
}
