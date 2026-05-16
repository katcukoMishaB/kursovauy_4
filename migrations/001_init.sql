CREATE TABLE groups (
    id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(64) NOT NULL UNIQUE
);
CREATE INDEX idx_groups_name ON groups(name);

INSERT INTO groups (name) VALUES
    ('ИП-21-1'), ('ИП-21-2'), ('ИП-22-1'), ('ИП-22-2'),
    ('ИВТ-21-1'), ('ИВТ-21-2'), ('ИВТ-22-1'),
    ('БИ-21-1'), ('БИ-22-1'),
    ('М-21-1'), ('М-22-1'),
    ('Преподаватели'), ('Сотрудники');

CREATE TABLE users (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name        VARCHAR(55)  NOT NULL,
    last_name         VARCHAR(55)  NOT NULL,
    email             VARCHAR(128) UNIQUE NOT NULL,
    password          VARCHAR(72)  NOT NULL,
    registration_date DATE         NOT NULL DEFAULT CURRENT_DATE,
    status            BOOLEAN      NOT NULL DEFAULT TRUE,
    group_id          UUID         NULL REFERENCES groups(id) ON DELETE SET NULL,
    user_type         VARCHAR(20)  NOT NULL DEFAULT 'student',
    CONSTRAINT users_user_type_chk CHECK (user_type IN ('student','teacher','staff'))
);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_group ON users(group_id);
CREATE INDEX idx_users_type  ON users(user_type);

CREATE TABLE user_roles (
    user_id        UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    is_participant BOOLEAN NOT NULL DEFAULT TRUE,
    is_organizer   BOOLEAN NOT NULL DEFAULT FALSE,
    is_admin       BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE organizer_requests (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    experience_description TEXT,
    resume_url             TEXT,
    request_type           VARCHAR(20) NOT NULL DEFAULT 'organizer',
    submission_date        DATE NOT NULL DEFAULT CURRENT_DATE,
    status                 VARCHAR(32) NOT NULL DEFAULT 'в рассмотрении',
    CONSTRAINT organizer_requests_type_chk CHECK (request_type IN ('organizer','teacher'))
);
CREATE INDEX idx_orgreq_user ON organizer_requests(user_id);
CREATE INDEX idx_orgreq_type ON organizer_requests(request_type);

CREATE TABLE project_categories (
    id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(128) NOT NULL UNIQUE
);

INSERT INTO project_categories (name) VALUES
    ('Технологии'),
    ('Образование'),
    ('Социальные проекты'),
    ('Искусство'),
    ('Наука'),
    ('Бизнес'),
    ('Экология'),
    ('Здоровье');

CREATE TABLE tag_catalog (
    id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(64) NOT NULL UNIQUE
);

CREATE TABLE projects (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organizer_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title             VARCHAR(128) NOT NULL,
    short_description VARCHAR(256),
    full_description  TEXT,
    goal_description  TEXT,
    image_url         TEXT,
    status            VARCHAR(32) NOT NULL DEFAULT 'активен',
    creation_date     DATE NOT NULL DEFAULT CURRENT_DATE,
    planned_end_date  DATE,
    completion_date   DATE
);
CREATE INDEX idx_projects_organizer ON projects(organizer_id);
CREATE INDEX idx_projects_status    ON projects(status);

CREATE TABLE project_category_links (
    project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES project_categories(id) ON DELETE CASCADE,
    PRIMARY KEY (project_id, category_id)
);
CREATE INDEX idx_pc_links_project  ON project_category_links(project_id);
CREATE INDEX idx_pc_links_category ON project_category_links(category_id);

CREATE TABLE project_tags (
    project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    tag_id      UUID NOT NULL REFERENCES tag_catalog(id) ON DELETE CASCADE,
    is_required BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (project_id, tag_id)
);
CREATE INDEX idx_project_tags_tag      ON project_tags(tag_id);
CREATE INDEX idx_project_tags_required ON project_tags(project_id, is_required);

CREATE TABLE project_goals (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id    UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    title         VARCHAR(255) NOT NULL,
    description   TEXT,
    is_achieved   BOOLEAN NOT NULL DEFAULT FALSE,
    creation_date DATE NOT NULL DEFAULT CURRENT_DATE,
    achieved_date DATE
);
CREATE INDEX idx_project_goals_project ON project_goals(project_id);

CREATE TABLE project_participation_requests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    comment         TEXT,
    resume_url      VARCHAR(256),
    submission_date DATE NOT NULL DEFAULT CURRENT_DATE,
    status          VARCHAR(32) NOT NULL DEFAULT 'в рассмотрении',
    UNIQUE (project_id, user_id)
);
CREATE INDEX idx_ppr_project ON project_participation_requests(project_id);
CREATE INDEX idx_ppr_user    ON project_participation_requests(user_id);

CREATE TABLE project_participations (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role       VARCHAR(32) NOT NULL DEFAULT 'участник',
    join_date  DATE NOT NULL DEFAULT CURRENT_DATE,
    UNIQUE (project_id, user_id),
    CONSTRAINT project_participations_role_chk
        CHECK (role IN ('участник','заместитель','руководитель'))
);
CREATE INDEX idx_pp_project ON project_participations(project_id);
CREATE INDEX idx_pp_user    ON project_participations(user_id);
CREATE UNIQUE INDEX uniq_project_leader
    ON project_participations(project_id) WHERE role = 'руководитель';

CREATE TABLE project_invitations (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id         UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    sender_id          UUID NOT NULL REFERENCES users(id)    ON DELETE CASCADE,
    recipient_user_id  UUID NULL     REFERENCES users(id)    ON DELETE CASCADE,
    recipient_email    VARCHAR(255) NULL,
    recipient_group_id UUID NULL     REFERENCES groups(id)   ON DELETE SET NULL,
    message            TEXT NOT NULL DEFAULT '',
    status             VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at         TIMESTAMP NOT NULL DEFAULT NOW(),
    responded_at       TIMESTAMP NULL,
    CONSTRAINT invitations_status_chk
        CHECK (status IN ('pending','accepted','rejected','cancelled')),
    CONSTRAINT invitations_recipient_chk
        CHECK (recipient_user_id IS NOT NULL OR recipient_email IS NOT NULL)
);
CREATE INDEX idx_inv_project   ON project_invitations(project_id);
CREATE INDEX idx_inv_recipient ON project_invitations(recipient_user_id);
CREATE INDEX idx_inv_email     ON project_invitations(recipient_email);
CREATE INDEX idx_inv_status    ON project_invitations(status);
CREATE INDEX idx_inv_group     ON project_invitations(recipient_group_id);


CREATE TABLE project_tasks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    assigned_to     UUID REFERENCES users(id) ON DELETE SET NULL,
    title           VARCHAR(128) NOT NULL,
    description     TEXT,
    status          VARCHAR(32) NOT NULL DEFAULT 'новая',
    priority        VARCHAR(20) NOT NULL DEFAULT 'средний',
    difficulty      INTEGER     NOT NULL DEFAULT 3,
    due_date        DATE,
    creation_date   DATE NOT NULL DEFAULT CURRENT_DATE,
    completion_date DATE,
    quality_rating  INTEGER,
    attachment_url  TEXT,
    CONSTRAINT project_tasks_status_chk
        CHECK (status IN ('новая','в работе','на проверке','завершена')),
    CONSTRAINT project_tasks_priority_chk
        CHECK (priority IN ('низкий','средний','высокий')),
    CONSTRAINT project_tasks_difficulty_chk
        CHECK (difficulty BETWEEN 1 AND 5),
    CONSTRAINT project_tasks_quality_chk
        CHECK (quality_rating IS NULL OR quality_rating BETWEEN 1 AND 5)
);
CREATE INDEX idx_project_tasks_project  ON project_tasks(project_id);
CREATE INDEX idx_project_tasks_assigned ON project_tasks(assigned_to);
CREATE INDEX idx_project_tasks_status   ON project_tasks(status);

CREATE TABLE project_task_assignees (
    task_id UUID NOT NULL REFERENCES project_tasks(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (task_id, user_id)
);
CREATE INDEX idx_task_assignees_user ON project_task_assignees(user_id);

CREATE TABLE task_comments (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id          UUID NOT NULL REFERENCES project_tasks(id) ON DELETE CASCADE,
    user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content          TEXT NOT NULL,
    publication_date DATE NOT NULL DEFAULT CURRENT_DATE
);
CREATE INDEX idx_task_comments_task ON task_comments(task_id);

CREATE TABLE project_chats (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id    UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name          VARCHAR(128) NOT NULL,
    creation_date DATE NOT NULL DEFAULT CURRENT_DATE
);
CREATE INDEX idx_project_chats_project ON project_chats(project_id);

CREATE TABLE chat_messages (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_id      UUID NOT NULL REFERENCES project_chats(id) ON DELETE CASCADE,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message_text TEXT NOT NULL,
    sent_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_chat_messages_chat ON chat_messages(chat_id);
CREATE INDEX idx_chat_messages_sent ON chat_messages(sent_at);

CREATE TABLE user_skills (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tag_id  UUID NOT NULL REFERENCES tag_catalog(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, tag_id)
);
CREATE INDEX idx_user_skills_tag ON user_skills(tag_id);

CREATE TABLE user_interests (
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES project_categories(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, category_id)
);

CREATE TABLE activity_log (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id  UUID REFERENCES projects(id) ON DELETE CASCADE,
    task_id     UUID REFERENCES project_tasks(id) ON DELETE CASCADE,
    action      VARCHAR(48) NOT NULL,
    occurred_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_activity_log_user    ON activity_log(user_id);
CREATE INDEX idx_activity_log_project ON activity_log(project_id);
CREATE INDEX idx_activity_log_at      ON activity_log(occurred_at);
