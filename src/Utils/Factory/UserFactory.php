<?php

namespace App\Utils\Factory;

use App\Entity\User;
use League\OAuth2\Client\Provider\FacebookUser;
use League\OAuth2\Client\Provider\GoogleUser;

/**
 * @TODO Соединить createFromFacebookUser и createFromGoogleUser. То что сейчас это бред полный
 */
class UserFactory
{
    /**
     *
     * @param FacebookUser $facebookUser
     *
     * @return User
     */
    public static function createFromFacebookUser(FacebookUser $facebookUser): User
    {
        $user = new User();
        $user->setEmail($facebookUser->getEmail());
        $user->setFullName($facebookUser->getName());
        $user->setFacebookId($facebookUser->getId());

        return $user;
    }

    /**
     *
     *
     * @param GoogleUser $googleUser
     *
     * @return User
     */
    public static function createFromGoogleUser(GoogleUser $googleUser): User
    {
        $user = new User();
        $user->setEmail($googleUser->getEmail());
        $user->setFullName($googleUser->getName());
        $user->setGoogleId($googleUser->getId());

        return $user;
    }
}
