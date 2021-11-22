USE shotify_db;

CREATE TABLE users (
	id int NOT NULL primary key AUTO_INCREMENT,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	username VARCHAR(30) NOT NULL,
	role enum('ADMIN', 'MODERATOR', 'USER') NOT NULL,
	bio VARCHAR(400) NOT NULL,
	avatar VARCHAR(300) NOT NULL,
	email VARCHAR(80) NOT NULL,
	password VARCHAR(500) NOT NULL
);

CREATE TABLE images (
	id int NOT NULL PRIMARY KEY AUTO_INCREMENT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
	url VARCHAR(500) NOT NULL,
	description VARCHAR(400) NOT NULL,
	user_id int NOT NULL,
	title VARCHAR(100) NOT NULL,
	price double NOT NULL,
	forSale Boolean NOT NULL,
	private Boolean NOT NULL
);

CREATE TABLE sales (
	id int NOT NULL PRIMARY KEY AUTO_INCREMENT,
	image_id int NOT NULL,
	buyer_id int NOT NULL,
	seller_id int NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    price double NOT NULL,
	UNIQUE(image_id, buyer_id, seller_id),
    CONSTRAINT CHK_IDs CHECK(buyer_id != seller_id)
);


CREATE TABLE labels (
	id int NOT NULL PRIMARY KEY  AUTO_INCREMENT,
	tag VARCHAR(25) NOT NULL,
    image_id int NOT NULL,
    UNIQUE(tag, image_id)
);


ALTER TABLE images ADD CONSTRAINT image_user_fkey FOREIGN KEY (user_id) REFERENCES users(id);


ALTER TABLE sales ADD CONSTRAINT sale_image_fkey FOREIGN KEY (image_id) REFERENCES images(id);
ALTER TABLE sales ADD CONSTRAINT sale_seller_fkey FOREIGN KEY (seller_id) REFERENCES users(id);
ALTER TABLE sales ADD CONSTRAINT sale_buyer_fkey FOREIGN KEY (buyer_id) REFERENCES users(id);


ALTER TABLE labels ADD CONSTRAINT label_image_fkey FOREIGN KEY (image_id) REFERENCES images(id) ON DELETE CASCADE;
