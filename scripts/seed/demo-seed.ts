/*
 * Demo seed for the fictional student professional organization
 * "Студенческая профессиональная лига СТУ".
 *
 * Bootstrap note:
 * A completely fresh database has no root division, no roles, and no permission-bearing
 * admin membership. Protected API mutations require those permissions, so the script
 * uses direct DB writes only to create/update:
 *   - the demo admin user
 *   - the single root division
 *   - the first SystemAdmin role, position, and membership
 *
 * After that bootstrap, organization data is created through the HTTP API so normal
 * authorization, validation, and event-log behavior are exercised.
 *
 * Run from activist_base after the stack is up:
 *   .\e2e\node_modules\.bin\tsc.cmd --target ES2022 --module NodeNext --moduleResolution NodeNext --types node --typeRoots .\e2e\node_modules\@types --skipLibCheck --outDir .\demo-seed\dist .\demo-seed\seed-demo.ts
 *   node .\demo-seed\dist\seed-demo.js
 */

import { spawnSync } from "node:child_process";

type Json = Record<string, unknown>;

type Link = {
  platform: string;
  value: string;
};

type RolePermission = {
  code: string;
  scope: string;
};

type Role = {
  id: string;
  name: string;
  permissions?: RolePermission[];
};

type DivisionNode = {
  id: string;
  parent_id?: string;
  short_name: string;
  full_name?: string;
  description?: string;
  regulation_url?: string;
  media_links?: Link[];
  is_archived?: boolean;
  children?: DivisionNode[];
};

type DivisionInput = {
  key: string;
  parentKey: string;
  shortName: string;
  fullName: string;
  description: string;
};

type Position = {
  id: string;
  title: string;
  role_id: string;
  division_id: string;
};

type Person = {
  key: string;
  login: string;
  password: string;
  firstName: string;
  lastName: string;
  middleName: string;
  gradebookNumber: string;
  groupNumber: string;
  institute: string;
  birthDate: string;
  phone: string;
  about: string;
  socialLinks: Link[];
};

type Assignment = {
  personKey: string;
  divisionKey: string;
  positionTitle: string;
  roleName: string;
  maxCount?: number;
};

const apiBaseUrl = normalizeApiBaseUrl(process.env.DEMO_API_BASE_URL || process.env.E2E_API_BASE_URL || "http://localhost:8080/api/v1");
const pgContainer = process.env.DEMO_PG_CONTAINER || "activist-base-pg";
const pgUser = process.env.DEMO_PG_USER || "postgres";
const pgDatabase = process.env.DEMO_PG_DATABASE || "activist_base";

const adminLogin = "demo_admin";
const adminPassword = "admin";
const adminId = "demo-admin";
const adminPasswordHash =
  "$argon2id$v=19$m=65536,t=3,p=2$AjrRLbbUcdKfNdY3qu7mDg$FtWDu4fNnu7/K2Y+gR/bkh+q8pKExAT69xCooT/aCOU";

const systemRoleName = "Администратор платформы Лиги";
const rootFullName = "Студенческая профессиональная лига Северного технологического университета";
const rootShortName = "Профлига СТУ";

const roles: Array<{ name: string; permissions: RolePermission[] }> = [
  {
    name: systemRoleName,
    permissions: [perm("system_admin", "current_and_descendants")]
  },
  {
    name: "Председатель объединения",
    permissions: [
      perm("can_add_member", "current_division"),
      perm("can_remove_member", "current_division"),
      perm("can_assign_position", "current_division"),
      perm("can_edit_division", "current_division"),
      perm("can_manage_positions", "current_division"),
      perm("can_create_subdivision", "current_and_descendants"),
      perm("can_archive_division", "current_and_descendants"),
      perm("can_view_contacts", "current_division"),
      perm("can_view_audit_log", "current_and_descendants")
    ]
  },
  {
    name: "Заместитель председателя",
    permissions: [
      perm("can_add_member", "current_division"),
      perm("can_remove_member", "current_division"),
      perm("can_assign_position", "current_division"),
      perm("can_edit_division", "current_division"),
      perm("can_manage_positions", "current_division"),
      perm("can_create_subdivision", "current_and_descendants"),
      perm("can_archive_division", "current_and_descendants"),
      perm("can_view_contacts", "current_division")
    ]
  },
  {
    name: "Куратор направления",
    permissions: [
      perm("can_add_member", "current_division"),
      perm("can_assign_position", "current_division"),
      perm("can_manage_positions", "current_division"),
      perm("can_create_subdivision", "current_and_descendants"),
      perm("can_view_contacts", "current_division")
    ]
  },
  {
    name: "Проектный координатор",
    permissions: [
      perm("can_add_member", "current_division"),
      perm("can_remove_member", "current_division"),
      perm("can_assign_position", "current_division"),
      perm("can_view_contacts", "current_division")
    ]
  },
  {
    name: "Медиаволонтер",
    permissions: [
      perm("can_view_contacts", "current_division"),
      perm("can_edit_self_profile", "self")
    ]
  },
  {
    name: "Участник секции",
    permissions: [perm("can_edit_self_profile", "self")]
  }
];

const divisions: DivisionInput[] = [
  div("presidium", "root", "Президиум Лиги", "Президиум Студенческой профессиональной лиги", "Выборный руководящий состав: повестка семестра, связь с администрацией университета и координация направлений."),
  div("secretariat", "root", "Секретариат и приемная", "Секретариат и студенческая приемная", "Регистрация активистов, протоколы встреч, прием заявок от студентов и сопровождение внутренних процессов."),
  div("records", "secretariat", "Документооборот и протоколы", "Документооборот и протоколы", "Повестки заседаний, решения президиума, списки участников и учет заявок."),
  div("volunteer-reg", "secretariat", "Волонтерская регистрация", "Волонтерская регистрация", "Онбординг новых участников, сбор контактов и распределение первичных задач."),
  div("education", "root", "Учебный комитет", "Учебный комитет", "Наставничество, карьерные маршруты, кейс-чемпионаты и помощь студентам в профессиональном развитии."),
  div("mentoring", "education", "Наставничество первокурсников", "Наставничество первокурсников", "Пары наставник-первокурсник, разбор учебных планов и адаптационные встречи."),
  div("career", "education", "Карьерные треки", "Карьерные треки", "Стажировки, встречи с выпускниками, CV-разборы и подготовка к собеседованиям."),
  div("cases", "education", "Команда кейс-чемпионатов", "Команда кейс-чемпионатов", "Сбор команд, тренировки презентаций и подготовка к межвузовским кейсам."),
  div("science", "root", "Научно-проектный отдел", "Научно-проектный отдел", "Студенческие исследования, стартапы, хакатоны и проектные лаборатории."),
  div("student-research", "science", "Студенческие исследования", "Студенческие исследования", "Постерные сессии, сбор научных групп и помощь с заявками на конференции."),
  div("startup-lab", "science", "Стартап-лаборатория", "Стартап-лаборатория", "Проверка идей, трекшн-встречи, питчи и связь с университетским бизнес-инкубатором."),
  div("hackathons", "science", "Инженерные хакатоны", "Инженерные хакатоны", "Технические соревнования, командообразование и воркшопы перед хакатонами."),
  div("clubs", "root", "Профессиональные клубы", "Профессиональные клубы", "Предметные сообщества по направлениям подготовки и профессиональным интересам студентов."),
  div("it-club", "clubs", "IT-клуб", "IT-клуб", "Встречи разработчиков, code review, пет-проекты и цифровые сервисы Лиги."),
  div("finance-club", "clubs", "Финансовый клуб", "Финансовый клуб", "Разбор рынков, финансовые кейсы, студенческие инвестиционные игры и встречи с экспертами."),
  div("law-club", "clubs", "Юридический клуб", "Юридический клуб", "Муткорты, правовые разборы, студенческие консультации и подготовка юридических мероприятий."),
  div("media", "root", "Медиацентр", "Медиацентр", "Новости Лиги, анонсы событий, фото, видео и редакционный календарь."),
  div("newsroom", "media", "Редакция новостей", "Редакция новостей", "Тексты, интервью, посты, рассылки и публикации на студенческих площадках."),
  div("photo-video", "media", "Фото- и видеогруппа", "Фото- и видеогруппа", "Съемка мероприятий, монтаж роликов и архив визуальных материалов."),
  div("culture", "root", "Культурно-массовый отдел", "Культурно-массовый отдел", "Кампусные мероприятия, творческие вечера, сцена и дизайн событий."),
  div("events", "culture", "Событийная группа", "Событийная группа", "Организация квизов, форумов, вечеров знакомств и календаря кампусных событий."),
  div("stage-design", "culture", "Дизайн и сцена", "Дизайн и сцена", "Визуальное оформление, сцена, мерч и навигация на мероприятиях."),
  div("sports", "root", "Спортивный отдел", "Спортивный отдел", "Факультетские турниры, студенческие сборные, киберспорт и массовые активности."),
  div("faculty-cups", "sports", "Турниры факультетов", "Турниры факультетов", "Лиги факультетов, заявки команд, сетки турниров и спортивные волонтеры."),
  div("esports", "sports", "Киберспорт", "Киберспорт", "Кибертурниры, стримы, комментаторы и студенческие команды."),
  div("student-legal", "root", "Правовая помощь студентам", "Правовая помощь студентам", "Помощь с учебными заявлениями, общежитием, стипендиями и правилами университета."),
  div("study-appeals", "student-legal", "Консультации по учебным вопросам", "Консультации по учебным вопросам", "Академические справки, пересдачи, индивидуальные планы и коммуникация с учебными офисами."),
  div("dorm-appeals", "student-legal", "Жилищно-бытовые обращения", "Жилищно-бытовые обращения", "Вопросы общежитий, бытовых условий и маршрутизация обращений к ответственным службам."),
  div("international", "root", "Международные связи", "Международные связи", "Поддержка иностранных студентов, разговорные клубы и межкультурные проекты."),
  div("buddy", "international", "Buddy program", "Buddy program", "Наставники для иностранных студентов, встречи адаптации и помощь с кампусной навигацией."),
  div("language-clubs", "international", "Разговорные клубы", "Разговорные клубы", "Английский, испанский и русский разговорные клубы для студентов разных программ.")
];

const people: Person[] = [
  person("melnikova", "demo_melnikova", "Дарья", "Мельникова", "Андреевна", "STU-0001", "4 курс", "Факультет компьютерных наук", "2003-05-14", "+7 916 401-12-10", "4 курс, ФКН. Председатель Лиги; ведет президиум и переговоры с деканатами.", "darya.melnikova"),
  person("karelin", "demo_karelin", "Иван", "Карелин", "Сергеевич", "STU-0002", "5 курс", "Инженерный факультет", "2002-11-03", "+7 915 204-44-19", "5 курс, инженерный факультет. Заместитель по проектам и календарю крупных инициатив.", "ivan.karelin"),
  person("safonova", "demo_safonova", "Алина", "Сафонова", "Игоревна", "STU-0003", "3 курс", "Факультет менеджмента", "2004-02-27", "+7 926 115-08-33", "3 курс, менеджмент. Ведет секретариат, протоколы встреч и регистрацию активистов.", "alina.safonova"),
  person("gromov", "demo_gromov", "Никита", "Громов", "Павлович", "STU-0004", "2 курс", "Прикладная информатика", "2005-09-18", "+7 903 774-18-05", "2 курс, прикладная информатика. Помогает первокурсникам с учебными маршрутами.", "nikita.gromov"),
  person("ustinova", "demo_ustinova", "Полина", "Устинова", "Максимовна", "STU-0005", "4 курс", "Экономический факультет", "2003-12-06", "+7 925 663-79-40", "4 курс, экономика. Собирает карьерные встречи, стажировки и встречи с выпускниками.", "polina.ustinova"),
  person("baranov", "demo_baranov", "Артем", "Баранов", "Денисович", "STU-0006", "3 курс", "Бизнес-информатика", "2004-07-21", "+7 985 221-62-18", "3 курс, бизнес-информатика. Готовит команды к кейс-чемпионатам.", "artem.baranov"),
  person("romanova", "demo_romanova", "Ксения", "Романова", "Олеговна", "STU-0007", "5 курс", "Биотехнологический факультет", "2001-10-30", "+7 921 775-31-09", "5 курс, биотех. Курирует студенческие исследования и постерные сессии.", "ksenia.romanova"),
  person("isaev", "demo_isaev", "Михаил", "Исаев", "Кириллович", "STU-0008", "4 курс", "Программная инженерия", "2003-03-12", "+7 920 332-14-77", "4 курс, программная инженерия. Организует хакатоны и технические воркшопы.", "mikhail.isaev"),
  person("zhukova", "demo_zhukova", "Варвара", "Жукова", "Алексеевна", "STU-0009", "2 курс", "Школа дизайна", "2005-08-25", "+7 916 707-88-12", "2 курс, дизайн. Ведет визуальный стиль мероприятий и сцены.", "varvara.zhukova"),
  person("tarasov", "demo_tarasov", "Егор", "Тарасов", "Викторович", "STU-0010", "3 курс", "Факультет журналистики", "2004-01-17", "+7 916 809-20-71", "3 курс, журналистика. Главный редактор студенческих новостей Лиги.", "egor.tarasov"),
  person("lazareva", "demo_lazareva", "Софья", "Лазарева", "Романовна", "STU-0011", "2 курс", "Филологический факультет", "2005-06-02", "+7 926 904-61-55", "2 курс, филология. Пишет новости, интервью и анонсы мероприятий.", "sofia.lazareva"),
  person("orlov", "demo_orlov", "Максим", "Орлов", "Ильич", "STU-0012", "4 курс", "Юридический факультет", "2003-04-09", "+7 977 301-45-64", "4 курс, юриспруденция. Помогает студентам с учебными заявлениями и правилами проживания.", "maxim.orlov"),
  person("nazarova", "demo_nazarova", "Елизавета", "Назарова", "Дмитриевна", "STU-0013", "3 курс", "Международные отношения", "2004-09-11", "+7 916 512-36-88", "3 курс, международные отношения. Ведет buddy program для иностранных студентов.", "elizaveta.nazarova"),
  person("kolesnikov", "demo_kolesnikov", "Тимур", "Колесников", "Артурович", "STU-0014", "2 курс", "Спортивный менеджмент", "2005-02-05", "+7 985 114-97-22", "2 курс, спортменеджмент. Организует турниры факультетов и сборные команды.", "timur.kolesnikov"),
  person("vinogradova", "demo_vinogradova", "Мария", "Виноградова", "Павловна", "STU-0015", "3 курс", "Программная инженерия", "2004-12-20", "+7 925 410-58-90", "3 курс, программная инженерия. Администрирует платформу и помогает с доступами.", "maria.vinogradova"),
  person("smirnov", "demo_smirnov", "Данила", "Смирнов", "Евгеньевич", "STU-0016", "1 курс", "Экономический факультет", "2006-03-28", "+7 916 305-74-11", "1 курс, экономика. Активист финансового клуба, помогает с регистрацией участников.", "danila.smirnov")
];

const assignments: Assignment[] = [
  { personKey: "melnikova", divisionKey: "root", positionTitle: "Председатель Студенческой профессиональной лиги", roleName: "Председатель объединения", maxCount: 1 },
  { personKey: "karelin", divisionKey: "presidium", positionTitle: "Заместитель председателя по проектам", roleName: "Заместитель председателя", maxCount: 1 },
  { personKey: "safonova", divisionKey: "secretariat", positionTitle: "Руководитель секретариата", roleName: "Куратор направления", maxCount: 1 },
  { personKey: "gromov", divisionKey: "mentoring", positionTitle: "Куратор наставничества", roleName: "Проектный координатор", maxCount: 1 },
  { personKey: "ustinova", divisionKey: "career", positionTitle: "Координатор карьерных треков", roleName: "Куратор направления", maxCount: 1 },
  { personKey: "baranov", divisionKey: "cases", positionTitle: "Тренер кейс-команд", roleName: "Проектный координатор", maxCount: 1 },
  { personKey: "romanova", divisionKey: "science", positionTitle: "Руководитель научно-проектного отдела", roleName: "Заместитель председателя", maxCount: 1 },
  { personKey: "isaev", divisionKey: "hackathons", positionTitle: "Организатор инженерных хакатонов", roleName: "Куратор направления", maxCount: 1 },
  { personKey: "zhukova", divisionKey: "stage-design", positionTitle: "Координатор дизайна и сцены", roleName: "Медиаволонтер", maxCount: 1 },
  { personKey: "tarasov", divisionKey: "media", positionTitle: "Главный редактор медиацентра", roleName: "Куратор направления", maxCount: 1 },
  { personKey: "lazareva", divisionKey: "newsroom", positionTitle: "Редактор новостей", roleName: "Медиаволонтер", maxCount: 1 },
  { personKey: "orlov", divisionKey: "student-legal", positionTitle: "Координатор правовой помощи", roleName: "Куратор направления", maxCount: 1 },
  { personKey: "nazarova", divisionKey: "buddy", positionTitle: "Координатор Buddy program", roleName: "Проектный координатор", maxCount: 1 },
  { personKey: "kolesnikov", divisionKey: "sports", positionTitle: "Руководитель спортивного отдела", roleName: "Куратор направления", maxCount: 1 },
  { personKey: "vinogradova", divisionKey: "it-club", positionTitle: "Администратор цифровой платформы", roleName: systemRoleName, maxCount: 1 },
  { personKey: "smirnov", divisionKey: "finance-club", positionTitle: "Активист финансового клуба", roleName: "Участник секции" }
];

function perm(code: string, scope: string): RolePermission {
  return { code, scope };
}

function div(key: string, parentKey: string, shortName: string, fullName: string, description: string): DivisionInput {
  return { key, parentKey, shortName, fullName, description };
}

function person(
  key: string,
  login: string,
  firstName: string,
  lastName: string,
  middleName: string,
  gradebookNumber: string,
  groupNumber: string,
  institute: string,
  birthDate: string,
  phone: string,
  about: string,
  emailLocal: string
): Person {
  return {
    key,
    login,
    password: "demo2026",
    firstName,
    lastName,
    middleName,
    gradebookNumber,
    groupNumber,
    institute,
    birthDate,
    phone,
    about: `${about} Университетская почта: ${emailLocal}@students.stu.ru.`,
    socialLinks: [
      { platform: "telegram", value: `https://t.me/stu_${emailLocal.replace(/\./g, "_")}` },
      { platform: "email", value: `${emailLocal}@students.stu.ru` }
    ]
  };
}

function normalizeApiBaseUrl(value: string): string {
  const trimmed = value.trim().replace(/\/+$/, "");
  return trimmed.endsWith("/api/v1") ? trimmed : `${trimmed}/api/v1`;
}

function sql(value: string): string {
  return `'${value.replace(/'/g, "''")}'`;
}

function runPsql(input: string): string {
  const result = spawnSync("docker", ["exec", "-i", pgContainer, "psql", "-U", pgUser, "-d", pgDatabase, "-v", "ON_ERROR_STOP=1", "-At", "-F", "\t"], {
    input,
    encoding: "utf8",
    maxBuffer: 10 * 1024 * 1024
  });
  if (result.error) {
    throw result.error;
  }
  if (result.status !== 0) {
    throw new Error(`psql failed:\n${result.stderr || result.stdout}`);
  }
  return result.stdout.trim();
}

function queryOne(sqlText: string): string[] {
  const output = runPsql(sqlText);
  const firstLine = output.split(/\r?\n/).find((line) => line.trim().length > 0);
  if (!firstLine) {
    throw new Error(`expected one SQL row, got none for:\n${sqlText}`);
  }
  return firstLine.split("\t");
}

function bootstrapSql(): string {
  return `
WITH admin_upsert AS (
  INSERT INTO users (id, login, password_hash, first_name, last_name, middle_name, gradebook_number, group_number, institute, birth_date, phone, social_links, about)
  VALUES (
    ${sql(adminId)},
    ${sql(adminLogin)},
    ${sql(adminPasswordHash)},
    ${sql("Демо")},
    ${sql("Администратор")},
    ${sql("Системный")},
    ${sql("STU-ADMIN")},
    ${sql("Демо")},
    ${sql("Платформа Лиги")},
    DATE ${sql("2002-04-01")},
    ${sql("+7 900 000-00-01")},
    '[]'::jsonb,
    ${sql("Техническая учетная запись для подготовки и проведения демонстрации студенческой платформы.")}
  )
  ON CONFLICT (login) DO UPDATE
  SET password_hash = EXCLUDED.password_hash,
      first_name = EXCLUDED.first_name,
      last_name = EXCLUDED.last_name,
      middle_name = EXCLUDED.middle_name,
      gradebook_number = EXCLUDED.gradebook_number,
      group_number = EXCLUDED.group_number,
      institute = EXCLUDED.institute,
      birth_date = EXCLUDED.birth_date,
      phone = EXCLUDED.phone,
      about = EXCLUDED.about,
      updated_at = now()
  RETURNING id
),
root_existing AS (
  SELECT id FROM divisions WHERE parent_id IS NULL LIMIT 1
),
root_inserted AS (
  INSERT INTO divisions (id, parent_id, short_name, full_name, description, regulation_url, media_links, is_archived)
  SELECT
    'demo-root',
    NULL,
    ${sql(rootShortName)},
    ${sql(rootFullName)},
    ${sql("Студенческое профессиональное объединение при Северном технологическом университете: карьерные треки, научные проекты, клубы, медиа и взаимопомощь студентов.")},
    ${sql("https://league.stu.example/regulations")},
    '[{"platform":"site","value":"https://league.stu.example"},{"platform":"telegram","value":"https://t.me/stu_prof_league"}]'::jsonb,
    false
  WHERE NOT EXISTS (SELECT 1 FROM root_existing)
  RETURNING id
),
root_selected AS (
  SELECT id FROM root_inserted
  UNION ALL
  SELECT id FROM root_existing
  LIMIT 1
),
role_upsert AS (
  INSERT INTO roles (id, name, permissions, created_at, updated_at)
  VALUES ('demo-role-system-admin', ${sql(systemRoleName)}, decode('01ff03', 'hex'), now(), now())
  ON CONFLICT (name) DO UPDATE
  SET permissions = EXCLUDED.permissions,
      updated_at = now()
  RETURNING id
),
position_upsert AS (
  INSERT INTO positions (id, title, role_id, division_id, max_count, is_archived)
  VALUES ('demo-pos-bootstrap-admin', ${sql("Демо-администратор платформы")}, (SELECT id FROM role_upsert), (SELECT id FROM root_selected), 1, false)
  ON CONFLICT (id) DO UPDATE
  SET title = EXCLUDED.title,
      role_id = EXCLUDED.role_id,
      division_id = EXCLUDED.division_id,
      max_count = EXCLUDED.max_count,
      is_archived = false
  RETURNING id
),
membership_upsert AS (
  INSERT INTO memberships (user_id, position_id)
  VALUES ((SELECT id FROM admin_upsert), (SELECT id FROM position_upsert))
  ON CONFLICT (user_id, position_id) DO NOTHING
)
SELECT (SELECT id FROM admin_upsert), (SELECT id FROM root_selected), (SELECT id FROM role_upsert), (SELECT id FROM position_upsert);
`;
}

class ApiClient {
  private cookie = "";
  private csrfToken = "";

  constructor(private readonly baseUrl: string) {}

  async login(login: string, password: string): Promise<void> {
    await this.request("POST", "/auth/login", { login, password }, 200, { csrf: false });
    const session = await this.request("GET", "/auth/session", undefined, 200, { csrf: false });
    this.csrfToken = String(session.csrf_token || "");
    if (!this.csrfToken) {
      throw new Error("login succeeded but /auth/session did not return csrf_token");
    }
  }

  async request(method: string, path: string, body?: unknown, expectedStatus = 200, options: { csrf?: boolean } = {}): Promise<any> {
    const headers: Record<string, string> = {
      Accept: "application/json"
    };
    if (body !== undefined) {
      headers["Content-Type"] = "application/json";
    }
    if (this.cookie) {
      headers.Cookie = this.cookie;
    }
    if (options.csrf !== false && ["POST", "PATCH", "PUT", "DELETE"].includes(method) && this.csrfToken) {
      headers["X-CSRF-Token"] = this.csrfToken;
    }

    const response = await fetch(`${this.baseUrl}${path}`, {
      method,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body)
    });

    const setCookie = response.headers.get("set-cookie");
    if (setCookie) {
      this.cookie = setCookie.split(";")[0];
    }

    const text = await response.text();
    const parsed = text.trim() ? JSON.parse(text) : null;
    if (response.status !== expectedStatus) {
      const message = parsed?.error?.message || parsed?.error?.code || text || response.statusText;
      throw new Error(`${method} ${path} expected ${expectedStatus}, got ${response.status}: ${message}`);
    }
    return parsed;
  }

  async requestAllowing(method: string, path: string, body: unknown, okStatuses: number[]): Promise<{ status: number; body: any }> {
    const headers: Record<string, string> = {
      Accept: "application/json"
    };
    if (body !== undefined) {
      headers["Content-Type"] = "application/json";
    }
    if (this.cookie) {
      headers.Cookie = this.cookie;
    }
    if (["POST", "PATCH", "PUT", "DELETE"].includes(method) && this.csrfToken) {
      headers["X-CSRF-Token"] = this.csrfToken;
    }

    const response = await fetch(`${this.baseUrl}${path}`, {
      method,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body)
    });
    const text = await response.text();
    const parsed = text.trim() ? JSON.parse(text) : null;
    if (!okStatuses.includes(response.status)) {
      const message = parsed?.error?.message || parsed?.error?.code || text || response.statusText;
      throw new Error(`${method} ${path} expected one of ${okStatuses.join(", ")}, got ${response.status}: ${message}`);
    }
    return { status: response.status, body: parsed };
  }
}

async function waitForApi(): Promise<void> {
  const healthUrl = apiBaseUrl.replace(/\/api\/v1$/, "/healthz");
  const deadline = Date.now() + 90_000;
  while (Date.now() < deadline) {
    try {
      const response = await fetch(healthUrl);
      if (response.status === 204) {
        return;
      }
    } catch {
      // Retry until backend is ready.
    }
    await sleep(1_000);
  }
  throw new Error(`API did not become ready at ${healthUrl}`);
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function ensureRole(api: ApiClient, name: string, permissions: RolePermission[]): Promise<Role> {
  const list = await api.request("GET", "/roles?limit=100", undefined, 200, { csrf: false });
  const existing = (list.items as Role[]).find((role) => role.name === name);
  if (!existing) {
    return api.request("POST", "/roles", { name, permissions }, 201);
  }
  if (!samePermissions(existing.permissions || [], permissions)) {
    return api.request("PATCH", `/roles/${encodeURIComponent(existing.id)}`, { permissions }, 200);
  }
  return existing;
}

function samePermissions(left: RolePermission[], right: RolePermission[]): boolean {
  const normalize = (items: RolePermission[]) => items.map((item) => `${item.code}:${item.scope}`).sort().join("|");
  return normalize(left) === normalize(right);
}

async function getTree(api: ApiClient): Promise<DivisionNode> {
  return api.request("GET", "/divisions/tree?depth=20", undefined, 200, { csrf: false });
}

function flattenTree(root: DivisionNode): DivisionNode[] {
  const result: DivisionNode[] = [];
  const stack = [root];
  while (stack.length > 0) {
    const node = stack.pop();
    if (!node) {
      continue;
    }
    result.push(node);
    for (const child of node.children || []) {
      stack.push(child);
    }
  }
  return result;
}

async function patchRoot(api: ApiClient, rootId: string): Promise<void> {
  await api.request("PATCH", `/divisions/${encodeURIComponent(rootId)}`, {
    short_name: rootShortName,
    full_name: rootFullName,
    description: "Студенческое профессиональное объединение при Северном технологическом университете: карьерные треки, научные проекты, клубы, медиа и взаимопомощь студентов.",
    regulation_url: "https://league.stu.example/regulations",
    media_links: [
      { platform: "site", value: "https://league.stu.example" },
      { platform: "telegram", value: "https://t.me/stu_prof_league" }
    ]
  }, 200);
}

async function ensureDivision(api: ApiClient, parentId: string, input: DivisionInput): Promise<DivisionNode> {
  const tree = await getTree(api);
  const existing = flattenTree(tree).find((node) => node.parent_id === parentId && !node.is_archived && node.short_name === input.shortName);
  if (existing) {
    if (existing.full_name !== input.fullName || existing.description !== input.description) {
      await api.request("PATCH", `/divisions/${encodeURIComponent(existing.id)}`, {
        short_name: input.shortName,
        full_name: input.fullName,
        description: input.description,
        regulation_url: "",
        media_links: []
      }, 200);
      return { ...existing, full_name: input.fullName, description: input.description };
    }
    return existing;
  }
  return api.request("POST", "/divisions", {
    parent_id: parentId,
    short_name: input.shortName,
    full_name: input.fullName,
    description: input.description,
    regulation_url: "",
    media_links: []
  }, 201);
}

async function ensureUser(adminApi: ApiClient, person: Person): Promise<string> {
  const registerPayload = {
    login: person.login,
    password: person.password,
    first_name: person.firstName,
    last_name: person.lastName,
    middle_name: person.middleName,
    gradebook_number: person.gradebookNumber,
    group_number: person.groupNumber,
    institute: person.institute,
    birth_date: person.birthDate,
    phone: person.phone,
    about: person.about
  };

  const publicApi = new ApiClient(apiBaseUrl);
  const registered = await publicApi.requestAllowing("POST", "/auth/register", registerPayload, [201, 400, 409, 422]);
  let userId = registered.body?.user?.id as string | undefined;
  if (!userId) {
    const search = await adminApi.request("GET", `/search/users?login=${encodeURIComponent(person.login)}&limit=10`, undefined, 200, { csrf: false });
    const exact = (search.items || []).find((item: any) => item.login === person.login);
    userId = exact?.id;
  }
  if (!userId) {
    throw new Error(`could not create or locate user ${person.login}`);
  }
  return userId;
}

async function ensurePosition(api: ApiClient, divisionId: string, title: string, roleId: string, maxCount?: number): Promise<Position> {
  const list = await api.request("GET", `/divisions/${encodeURIComponent(divisionId)}/positions`, undefined, 200, { csrf: false });
  const existing = (list.items as Position[]).find((position) => position.title === title);
  if (existing) {
    return existing;
  }
  const payload: Json = {
    title,
    role_id: roleId
  };
  if (maxCount !== undefined) {
    payload.max_count = maxCount;
  }
  return api.request("POST", `/divisions/${encodeURIComponent(divisionId)}/positions`, payload, 201);
}

async function ensureMembership(api: ApiClient, userId: string, positionId: string): Promise<void> {
  await api.requestAllowing("POST", "/memberships", {
    user_id: userId,
    position_id: positionId
  }, [201, 409]);
}

async function createArchiveActivity(api: ApiClient, divisionsByKey: Map<string, DivisionNode>, roleByName: Map<string, Role>, adminUserId: string): Promise<void> {
  const parent = divisionsByKey.get("science");
  const role = roleByName.get("Куратор направления");
  if (!parent || !role) {
    return;
  }
  const tempDivision = await api.request("POST", "/divisions", {
    parent_id: parent.id,
    short_name: "Демодень стартапов",
    full_name: "Временная группа: демодень студенческих стартапов",
    description: "Короткая рабочая группа для подготовки питч-сессии, закрытая после проведения демодня.",
    regulation_url: "",
    media_links: []
  }, 201);
  const tempPosition = await api.request("POST", `/divisions/${encodeURIComponent(tempDivision.id)}/positions`, {
    title: "Координатор демодня",
    role_id: role.id,
    max_count: 1
  }, 201);
  await ensureMembership(api, adminUserId, tempPosition.id);
  await api.request("POST", `/positions/${encodeURIComponent(tempPosition.id)}/archive`, {}, 200);
  await api.request("POST", `/divisions/${encodeURIComponent(tempDivision.id)}/archive`, {}, 200);
}

async function main(): Promise<void> {
  console.log(`[demo-seed] API: ${apiBaseUrl}`);
  console.log("[demo-seed] waiting for API readiness");
  await waitForApi();

  console.log("[demo-seed] bootstrapping admin/root/system-admin seat via DB");
  const [bootAdminId, rootId] = queryOne(bootstrapSql());
  if (!bootAdminId || !rootId) {
    throw new Error("bootstrap did not return admin/root ids");
  }

  const adminApi = new ApiClient(apiBaseUrl);
  await adminApi.login(adminLogin, adminPassword);
  await patchRoot(adminApi, rootId);

  console.log("[demo-seed] ensuring roles");
  const roleByName = new Map<string, Role>();
  for (const roleInput of roles) {
    const role = await ensureRole(adminApi, roleInput.name, roleInput.permissions);
    roleByName.set(role.name, role);
  }

  console.log("[demo-seed] ensuring division tree");
  const divisionsByKey = new Map<string, DivisionNode>();
  divisionsByKey.set("root", { id: rootId, short_name: rootShortName, full_name: rootFullName });
  for (const divisionInput of divisions) {
    const parent = divisionsByKey.get(divisionInput.parentKey);
    if (!parent) {
      throw new Error(`missing parent division key ${divisionInput.parentKey} for ${divisionInput.key}`);
    }
    const division = await ensureDivision(adminApi, parent.id, divisionInput);
    divisionsByKey.set(divisionInput.key, division);
  }

  console.log("[demo-seed] ensuring people and profiles");
  const userIdByPersonKey = new Map<string, string>();
  for (const personInput of people) {
    const userId = await ensureUser(adminApi, personInput);
    userIdByPersonKey.set(personInput.key, userId);
  }

  console.log("[demo-seed] ensuring positions and memberships");
  for (const assignment of assignments) {
    const userId = userIdByPersonKey.get(assignment.personKey);
    const divisionNode = divisionsByKey.get(assignment.divisionKey);
    const role = roleByName.get(assignment.roleName);
    if (!userId || !divisionNode || !role) {
      throw new Error(`assignment cannot be resolved: ${JSON.stringify(assignment)}`);
    }
    const position = await ensurePosition(adminApi, divisionNode.id, assignment.positionTitle, role.id, assignment.maxCount);
    await ensureMembership(adminApi, userId, position.id);
  }

  console.log("[demo-seed] adding extra archive/audit activity");
  await createArchiveActivity(adminApi, divisionsByKey, roleByName, bootAdminId);

  const eventLog = await adminApi.request("GET", "/eventlog?limit=5", undefined, 200, { csrf: false });
  const tree = await getTree(adminApi);
  console.log(`[demo-seed] complete: ${flattenTree(tree).length} active divisions, ${people.length} people, latest audit rows returned=${eventLog.items?.length || 0}`);
  console.log(`[demo-seed] login: ${adminLogin} / ${adminPassword}`);
}

main().catch((error) => {
  console.error(`[demo-seed] failed: ${error instanceof Error ? error.message : String(error)}`);
  process.exit(1);
});
