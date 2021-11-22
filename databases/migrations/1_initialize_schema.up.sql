CREATE DATABASE IF NOT EXISTS shotify_db;
USE shotify_db;

CREATE TABLE IF NOT EXISTS `images` (
  `id` int NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `url` varchar(500) NOT NULL,
  `description` varchar(400) NOT NULL,
  `user_id` int NOT NULL,
  `title` varchar(100) NOT NULL,
  `price` double NOT NULL,
  `forSale` tinyint(1) NOT NULL,
  `private` tinyint(1) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `image_user_fkey` (`user_id`),
  CONSTRAINT `image_user_fkey` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ;

CREATE TABLE IF NOT EXISTS `labels` (
  `id` int NOT NULL AUTO_INCREMENT,
  `tag` varchar(25) NOT NULL,
  `image_id` int NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `tag` (`tag`,`image_id`),
  KEY `label_image_fkey` (`image_id`),
  CONSTRAINT `label_image_fkey` FOREIGN KEY (`image_id`) REFERENCES `images` (`id`) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS `sales` (
  `id` int NOT NULL AUTO_INCREMENT,
  `image_id` int NOT NULL,
  `buyer_id` int NOT NULL,
  `seller_id` int NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `price` double NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `image_id` (`image_id`,`buyer_id`,`seller_id`),
  KEY `sale_seller_fkey` (`seller_id`),
  KEY `sale_buyer_fkey` (`buyer_id`),
  CONSTRAINT `sale_buyer_fkey` FOREIGN KEY (`buyer_id`) REFERENCES `users` (`id`),
  CONSTRAINT `sale_image_fkey` FOREIGN KEY (`image_id`) REFERENCES `images` (`id`),
  CONSTRAINT `sale_seller_fkey` FOREIGN KEY (`seller_id`) REFERENCES `users` (`id`),
  CONSTRAINT `CHK_IDs` CHECK ((`buyer_id` <> `seller_id`))
);

CREATE TABLE IF NOT EXISTS `users` (
  `id` int NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `username` varchar(30) NOT NULL,
  `role` enum('ADMIN','MODERATOR','USER') NOT NULL,
  `bio` varchar(400) NOT NULL,
  `avatar` varchar(300) NOT NULL,
  `email` varchar(80) NOT NULL,
  `password` varchar(500) NOT NULL,
  PRIMARY KEY (`id`)
);
