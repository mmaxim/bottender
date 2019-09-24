-- MySQL dump 10.13  Distrib 5.6.43, for osx10.14 (x86_64)
--
-- Host: localhost    Database: drinks
-- ------------------------------------------------------
-- Server version	5.6.43

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Current Database: `drinks`
--

CREATE DATABASE /*!32312 IF NOT EXISTS*/ `drinks` /*!40100 DEFAULT CHARACTER SET utf8 */;

USE `drinks`;

--
-- Table structure for table `drink_ingredients`
--

DROP TABLE IF EXISTS `drink_ingredients`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `drink_ingredients` (
  `drink_id` int(11) NOT NULL,
  `ingredient_id` int(11) NOT NULL,
  `amount` int(11) NOT NULL,
  KEY `drink_id` (`drink_id`),
  KEY `ingredient_id` (`ingredient_id`),
  CONSTRAINT `drink_ingredients_ibfk_1` FOREIGN KEY (`drink_id`) REFERENCES `drinks` (`id`),
  CONSTRAINT `drink_ingredients_ibfk_2` FOREIGN KEY (`ingredient_id`) REFERENCES `ingredient` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `drink_ingredients`
--

LOCK TABLES `drink_ingredients` WRITE;
/*!40000 ALTER TABLE `drink_ingredients` DISABLE KEYS */;
INSERT INTO `drink_ingredients` VALUES (1,2,100),(1,3,6),(1,4,1),(1,5,200),(2,8,300),(2,26,50),(3,1,200),(3,3,6),(3,25,1),(4,7,150),(4,6,50),(4,29,100),(4,17,35),(4,22,1),(3,30,1),(5,8,100),(5,14,100),(5,2,100),(5,30,1),(6,11,150),(6,6,75),(6,18,50),(6,30,1),(7,8,400),(7,12,25),(7,18,25),(7,3,6),(7,4,1),(8,9,150),(8,13,75),(8,17,50),(9,15,200),(9,6,100),(9,17,75),(9,22,1),(10,15,200),(10,13,100),(10,17,75),(10,22,1),(11,9,200),(11,12,25),(11,28,75),(11,17,30),(11,23,25),(11,22,1),(12,9,200),(12,17,100),(12,23,75),(12,22,1),(13,7,200),(13,17,100),(13,6,100),(13,22,1),(14,7,200),(14,32,500),(14,31,1),(15,15,250),(15,33,200),(15,28,400),(15,22,1),(16,1,200),(16,18,75),(16,23,75),(16,31,1),(16,4,1),(17,1,200),(17,13,75),(17,27,6),(18,8,100),(18,13,50),(18,26,100),(19,5,200),(19,34,50),(19,2,50),(19,3,1),(19,27,1),(20,7,150),(20,28,100),(20,29,400),(20,22,1),(21,1,200),(21,14,100),(21,2,100),(21,19,1),(22,15,200),(22,35,100),(22,18,75),(22,19,1),(23,9,150),(23,35,100),(23,36,200),(23,37,1),(24,8,200),(24,35,100),(24,6,100),(24,6,100),(24,27,3),(25,8,150),(25,13,100),(25,34,75),(25,17,75),(25,3,3),(26,8,150),(26,34,50),(26,2,100),(26,3,4),(27,38,100),(27,2,100),(27,1,100),(27,3,2),(27,22,1),(28,8,200),(28,23,100),(28,18,75),(28,33,300),(28,19,1),(28,4,1),(29,9,200),(29,10,200),(29,32,100),(29,39,200),(29,17,50),(29,23,50),(29,40,50),(29,4,1),(29,31,1);
/*!40000 ALTER TABLE `drink_ingredients` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `drinks`
--

DROP TABLE IF EXISTS `drinks`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `drinks` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL,
  `mixing` enum('shaken','stirred') DEFAULT NULL,
  `glass` enum('rocks','coupe','martini','hiball','hurricane') DEFAULT NULL,
  `serving` enum('rocks','up','neat') DEFAULT NULL,
  `notes` text,
  PRIMARY KEY (`id`),
  KEY `mixing` (`mixing`),
  KEY `glass` (`glass`),
  KEY `serving` (`serving`)
) ENGINE=InnoDB AUTO_INCREMENT=30 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `drinks`
--

LOCK TABLES `drinks` WRITE;
/*!40000 ALTER TABLE `drinks` DISABLE KEYS */;
INSERT INTO `drinks` VALUES (1,'Manhattan','stirred','coupe','up','Strong drink for whisky lovers. Stir for 30-45 seconds and strain.'),(2,'Martini','shaken','martini','up','Can take olives for \"dirtyness\", and more or less vermouth for \"dryness\"'),(3,'Old Fashioned','stirred','rocks','rocks','Muddle the sugar cube with help from bitters, stir with ice in the glass'),(4,'Cosmopolitan','shaken','martini','up','Easy on the lime, shake well'),(5,'Negroni','stirred','rocks','rocks','Stir in the glass, can add orange juice to cut Campari'),(6,'Sidecar','shaken','coupe','up','Can add sugar to sweeten it up'),(7,'Casino','shaken','coupe','up','Could consider stirring it'),(8,'Parisian Daiquiri','shaken','coupe','up','Can add sugar to make it sweeter'),(9,'Margarita','shaken','rocks','up','Can salt rim of glass'),(10,'Elderflower Margarita','shaken','rocks','up','Can salt rim of glass'),(11,'Hemingway Daiquiri','shaken','martini','up','Careful with maraschino, it is very strong'),(12,'Daiquiri','shaken','martini','up','Other variants might be better'),(13,'Kamikaze','shaken','martini','up','Margarita with vodka'),(14,'Screwdriver','stirred','hiball','rocks',''),(15,'Paloma','stirred','hiball','rocks',''),(16,'Whisky Sour','shaken','rocks','rocks','Easy on the lemon juice, or go hard if people like it'),(17,'Elderfashioned','stirred','rocks','rocks',''),(18,'Olivette','stirred','martini','up','Feature St Germain'),(19,'Greenpoint','stirred','coupe','up','Mess up your Manhattan with Chartreuse'),(20,'Seabreeze','stirred','hiball','rocks','Refreshing vodka cocktail'),(21,'Boulevardier','shaken','coupe','up','Can also stir this'),(22,'Blue Margarita','shaken','martini','up','Just swaps triple sec for blue curacao'),(23,'Blue Hawaiian','shaken','rocks','rocks','Can also serve in a wine glass'),(24,'Bluebird','shaken','martini','up',''),(25,'Lumiere','shaken','coupe','up',''),(26,'Bijou','stirred','martini','up',''),(27,'Blood and Bourbon','shaken','coupe','up',''),(28,'Tom Collins','shaken','hiball','rocks','Put everything but club soda into the shaker and give a quick shake. Then can take ice and put it back into a highball glass and top with club soda.'),(29,'Hurricane','shaken','hurricane','rocks','Can use some kind of passion fruit syrup to substitute for the juice.');
/*!40000 ALTER TABLE `drinks` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ingredient`
--

DROP TABLE IF EXISTS `ingredient`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `ingredient` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL,
  `desc` text,
  `category` enum('spirit','liqueur','aromatic','bitters','citrus','sugar','garnish') DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `category` (`category`)
) ENGINE=InnoDB AUTO_INCREMENT=41 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ingredient`
--

LOCK TABLES `ingredient` WRITE;
/*!40000 ALTER TABLE `ingredient` DISABLE KEYS */;
INSERT INTO `ingredient` VALUES (1,'Bourbon','American Whisky: Maker\'s Mark and others','spirit'),(2,'Sweet Vermouth','Fortified wine: Noilly Pratt, Carpano Antica, etc','aromatic'),(3,'Orange Bitters','Angostura orange bitters','bitters'),(4,'Maraschino Cherries','Fermented cherries','garnish'),(5,'Rye Whisky','Rye whisky, sharper than normal. Rittenhouse','spirit'),(6,'Triple Sec','Orange liqueur, Cointreau','liqueur'),(7,'Vodka','Neutral grain spirit','spirit'),(8,'Gin','Juniper infused spirit, Hendrik\'s, Tanqueray 10','spirit'),(9,'White Rum','Sweet spirit, Bacardi','spirit'),(10,'Dark Rum','Dark version of normal rum, Bacardi Black','spirit'),(11,'Cognac','Distilled wine, Hennessey','spirit'),(12,'Maraschino','Cherry liqueur, Luxardo','liqueur'),(13,'St Germain','Elderflower liqueur','liqueur'),(14,'Campari','Bitter red liqueur','liqueur'),(15,'Tequila','Mexican alcohol: Don Juio, Patron','spirit'),(16,'Aperol','Sweeter version of Campari','liqueur'),(17,'Lime Juice','Fresh squeezed','citrus'),(18,'Lemon Juice','Fresh squeezed','citrus'),(19,'Lemon Wedge','','garnish'),(20,'Lemon Peel','','garnish'),(21,'Lime Peel','','garnish'),(22,'Lime Wedge','','garnish'),(23,'Simple Syrup','Sugar in water','sugar'),(24,'Sugar','','sugar'),(25,'Sugar Cube','','sugar'),(26,'Dry Vermouth','Fortified wine: Noilly Pratt, used in martinis','aromatic'),(27,'Aromatic Bitters','Angostura normal bitters','bitters'),(28,'Grapefruit Juice','','citrus'),(29,'Cranberry Juice','','citrus'),(30,'Orange Peel','','garnish'),(31,'Orange Wedge','','garnish'),(32,'Orange Juice','','citrus'),(33,'Club Soda','','citrus'),(34,'Chartreuse','Green very strong liqueur','liqueur'),(35,'Blue Curacao','Make all your drinks blue, subs for triple sec in any recipe','liqueur'),(36,'Pineapple Juice','','citrus'),(37,'Pineapple Slice','','garnish'),(38,'Solerno','Bloord orange liqueur','liqueur'),(39,'Passion Fruit Juice','','citrus'),(40,'Grenadine','','citrus');
/*!40000 ALTER TABLE `ingredient` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ingredient_search_tags`
--

DROP TABLE IF EXISTS `ingredient_search_tags`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `ingredient_search_tags` (
  `ingredient_id` int(11) NOT NULL,
  `tag` varchar(100) DEFAULT NULL,
  KEY `ingredient_id` (`ingredient_id`),
  CONSTRAINT `ingredient_search_tags_ibfk_1` FOREIGN KEY (`ingredient_id`) REFERENCES `ingredient` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ingredient_search_tags`
--

LOCK TABLES `ingredient_search_tags` WRITE;
/*!40000 ALTER TABLE `ingredient_search_tags` DISABLE KEYS */;
/*!40000 ALTER TABLE `ingredient_search_tags` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2019-09-23 22:04:22
