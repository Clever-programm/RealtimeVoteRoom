CREATE TABLE polls (
    id SERIAL PRIMARY KEY,
    question TEXT NOT NULL
);

CREATE TABLE options (
    id SERIAL PRIMARY KEY,
    poll_id INT REFERENCES polls(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    votes INT DEFAULT 0
);

INSERT INTO polls (id, question) VALUES (1, 'Какой стек выбрать для реалтайм пет-проекта?');

INSERT INTO options (poll_id, text, votes) VALUES 
(1, 'Go + Flutter + Postgres', 10),
(1, 'Node.js + React + MongoDB', 3),
(1, 'Python + Vue + SQLite', 5);