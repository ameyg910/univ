CREATE TABLE users (
    id UUID PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    verified BOOLEAN DEFAULT FALSE,
    full_name TEXT,
    department TEXT,
    year TEXT,
    bio TEXT,
    profile_pic TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE otps (
    email TEXT PRIMARY KEY,
    code TEXT NOT NULL,
    expires_at TIMESTAMP
);

CREATE TABLE swipes (
    id SERIAL PRIMARY KEY,
    swiper_id UUID,
    swipee_id UUID,
    direction TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE matches (
    id SERIAL PRIMARY KEY,
    user1 UUID,
    user2 UUID,
    matched_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
