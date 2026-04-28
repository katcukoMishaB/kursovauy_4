CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    first_name VARCHAR(55) NOT NULL,
    last_name VARCHAR(55) NOT NULL,
    email VARCHAR(55) UNIQUE NOT NULL,
    phone VARCHAR(22),
    password VARCHAR(64) NOT NULL,
    registration_date DATE NOT NULL DEFAULT CURRENT_DATE,
    status BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE user_roles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    is_participant BOOLEAN NOT NULL DEFAULT true,
    is_organizer BOOLEAN NOT NULL DEFAULT false,
    is_admin BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE organizer_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    experience_description TEXT,
    submission_date DATE NOT NULL DEFAULT CURRENT_DATE,
    status VARCHAR(32) NOT NULL DEFAULT 'в рассмотрении',
    admin_comment TEXT
);

CREATE TABLE project_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(128) NOT NULL UNIQUE
);

CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organizer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id UUID REFERENCES project_categories(id) ON DELETE SET NULL,
    title VARCHAR(128) NOT NULL,
    short_description VARCHAR(256),
    full_description TEXT,
    status VARCHAR(32) NOT NULL DEFAULT 'активен',
    creation_date DATE NOT NULL DEFAULT CURRENT_DATE,
    completion_date DATE
);

CREATE TABLE project_participation_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    comment TEXT,
    resume_url VARCHAR(256),
    submission_date DATE NOT NULL DEFAULT CURRENT_DATE,
    status VARCHAR(32) NOT NULL DEFAULT 'в рассмотрении',
    UNIQUE(project_id, user_id)
);

CREATE TABLE project_participations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(32) NOT NULL DEFAULT 'участник',
    join_date DATE NOT NULL DEFAULT CURRENT_DATE,
    UNIQUE(project_id, user_id)
);

CREATE TABLE project_tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    assigned_to UUID REFERENCES users(id) ON DELETE SET NULL,
    title VARCHAR(128) NOT NULL,
    description TEXT,
    status VARCHAR(32) NOT NULL DEFAULT 'новая',
    creation_date DATE NOT NULL DEFAULT CURRENT_DATE
);

CREATE TABLE task_comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES project_tasks(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    publication_date DATE NOT NULL DEFAULT CURRENT_DATE
);

CREATE TABLE project_chats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(128) NOT NULL,
    creation_date DATE NOT NULL DEFAULT CURRENT_DATE
);

CREATE TABLE chat_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    chat_id UUID NOT NULL REFERENCES project_chats(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message_text TEXT NOT NULL,
    sent_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE project_tags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(128) NOT NULL,
    UNIQUE(project_id, name)
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_projects_organizer ON projects(organizer_id);
CREATE INDEX idx_projects_status ON projects(status);
CREATE INDEX idx_project_tasks_project ON project_tasks(project_id);
CREATE INDEX idx_project_tasks_assigned ON project_tasks(assigned_to);
CREATE INDEX idx_chat_messages_chat ON chat_messages(chat_id);
CREATE INDEX idx_project_participations_project ON project_participations(project_id);
CREATE INDEX idx_project_participations_user ON project_participations(user_id);

INSERT INTO project_categories (name) VALUES 
    ('Технологии'),
    ('Образование'),
    ('Социальные проекты'),
    ('Искусство'),
    ('Наука'),
    ('Бизнес'),
    ('Экология'),
    ('Здоровье');

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO crowdsourcing;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO crowdsourcing;
GRANT USAGE, CREATE ON SCHEMA public TO crowdsourcing;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO crowdsourcing;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO crowdsourcing;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE users (id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),first_name VARCHAR(55) NOT NULL,last_name VARCHAR(55) NOT NULL,email VARCHAR(55) UNIQUE NOT NULL,phone VARCHAR(22),password VARCHAR(64) NOT NULL,registration_date DATE NOT NULL DEFAULT CURRENT_DATE,status BOOLEAN NOT NULL DEFAULT true;
CREATE TABLE user_roles (user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,is_participant BOOLEAN NOT NULL DEFAULT true,is_organizer BOOLEAN NOT NULL DEFAULT false,is_admin BOOLEAN NOT NULL DEFAULT false);
CREATE TABLE organizer_requests (id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,experience_description TEXT,submission_date DATE NOT NULL DEFAULT CURRENT_DATE,status VARCHAR(32) NOT NULL DEFAULT 'в рассмотрении',admin_comment TEXT);
CREATE TABLE project_categories (id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),name VARCHAR(128) NOT NULL UNIQUE);
CREATE TABLE projects (id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),organizer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,category_id UUID REFERENCES project_categories(id) ON DELETE SET NULL,title VARCHAR(128) NOT NULL,short_description VARCHAR(256),full_description TEXT,status VARCHAR(32) NOT NULL DEFAULT 'активен',creation_date DATE NOT NULL DEFAULT CURRENT_DATE,completion_date DATE);
CREATE TABLE project_participation_requests (id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,comment TEXT,resume_url VARCHAR(256),submission_date DATE NOT NULL DEFAULT CURRENT_DATE,status VARCHAR(32) NOT NULL DEFAULT 'в рассмотрении',UNIQUE(project_id, user_id));
CREATE TABLE project_participations (id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,role VARCHAR(32) NOT NULL DEFAULT 'участник',join_date DATE NOT NULL DEFAULT CURRENT_DATE,UNIQUE(project_id, user_id));
CREATE TABLE project_tasks (id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,assigned_to UUID REFERENCES users(id) ON DELETE SET NULL,title VARCHAR(128) NOT NULL,description TEXT,status VARCHAR(32) NOT NULL DEFAULT 'новая',creation_date DATE NOT NULL DEFAULT CURRENT_DATE);
CREATE TABLE task_comments (id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),task_id UUID NOT NULL REFERENCES project_tasks(id) ON DELETE CASCADE,user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,content TEXT NOT NULL,publication_date DATE NOT NULL DEFAULT CURRENT_DATE);
CREATE TABLE project_chats (id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,name VARCHAR(128) NOT NULL,creation_date DATE NOT NULL DEFAULT CURRENT_DATE);
CREATE TABLE chat_messages (id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),chat_id UUID NOT NULL REFERENCES project_chats(id) ON DELETE CASCADE,user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,message_text TEXT NOT NULL,sent_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE project_tags (id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,name VARCHAR(128) NOT NULL,UNIQUE(project_id, name));
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_projects_organizer ON projects(organizer_id);
CREATE INDEX idx_projects_status ON projects(status);
CREATE INDEX idx_project_tasks_project ON project_tasks(project_id);
CREATE INDEX idx_project_tasks_assigned ON project_tasks(assigned_to);
CREATE INDEX idx_chat_messages_chat ON chat_messages(chat_id);
CREATE INDEX idx_project_participations_project ON project_participations(project_id);
CREATE INDEX idx_project_participations_user ON project_participations(user_id);
INSERT INTO project_categories (name) VALUES ('Технологии'),('Образование'),('Социальные проекты'),('Искусство'),('Наука'),('Бизнес'),('Экология'),('Здоровье');
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO crowdsourcing;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO crowdsourcing;
GRANT USAGE, CREATE ON SCHEMA public TO crowdsourcing;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO crowdsourcing;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO crowdsourcing;

