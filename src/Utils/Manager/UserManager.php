<?php

namespace App\Utils\Manager;

use App\Entity\User;
use App\Exception\Security\EmptyUserPlainPasswordException;
use Doctrine\ORM\EntityManagerInterface;
use Doctrine\Persistence\ObjectRepository;
use Symfony\Component\PasswordHasher\Hasher\UserPasswordHasherInterface;

class UserManager extends AbstractBaseManager
{
    private UserPasswordHasherInterface $passwordHasher;

    public function __construct(EntityManagerInterface $entityManager, UserPasswordHasherInterface $passwordHasher)
    {
        parent::__construct($entityManager);
        $this->passwordHasher = $passwordHasher;
    }

    /**
     * @return ObjectRepository
     */
    public function getRepository(): ObjectRepository
    {
        return $this->entityManager->getRepository(User::class);
    }

    /**
     * @param object $user
     * @return void
     */
    public function remove(object $user): void
    {
        /** @var User $user */
        $user->setIsDeleted(true);

        $this->save($user);
        $this->entityManager->flush();
    }

    /**
     *
     * @param User $user
     * @param string $plainPassword
     *
     * @return void
     */
    public function encodePassword(User $user, string $plainPassword): void
    {
        $plainPassword = trim($plainPassword);

        if (!$plainPassword) {
            throw new EmptyUserPlainPasswordException(message: 'Empty User password');
        }

        $user->setPassword($this->passwordHasher->hashPassword($user, $plainPassword));

    }
}