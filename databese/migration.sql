USE platform_engineer_clone;

DROP TABLE IF EXISTS `user`;
CREATE TABLE `user` (
                        `id` int NOT NULL AUTO_INCREMENT,
                        `name` varchar(255) NOT NULL,
                        `email` varchar(320) NOT NULL,
                        `password` varchar(255) NOT NULL,
                        `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        PRIMARY KEY (`id`),
                        UNIQUE KEY `user_email_uindex` (`email`),
                        UNIQUE KEY `user_name_uindex` (`name`)
);

DROP TABLE IF EXISTS `token`;
CREATE TABLE `token` (
                         `id` int NOT NULL AUTO_INCREMENT,
                         `key` varchar(12) NOT NULL,
                         `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                         `revoked` tinyint(1) NOT NULL DEFAULT '0',
                         `expired` tinyint(1) NOT NULL DEFAULT '0',
                         `created_by` int NOT NULL,
                         `expires_at` timestamp NOT NULL,
                         PRIMARY KEY (`id`),
                         UNIQUE KEY `token_name_uindex` (`key`),
                         KEY `token_user_id_fk` (`created_by`),
                         CONSTRAINT `token_user_id_fk` FOREIGN KEY (`created_by`) REFERENCES `user` (`id`)
);