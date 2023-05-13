<?php

namespace App\Form\Handler;

use App\Entity\User;
use App\Utils\Manager\UserManager;
use Symfony\Component\Form\Form;
use Symfony\Component\Form\FormInterface;
use Symfony\Component\PasswordHasher\Hasher\UserPasswordHasherInterface;

class UserFormHandler
{
    private UserManager $userManager;
    private UserPasswordHasherInterface $userPasswordHasher;

    public function __construct(UserManager $userManager, UserPasswordHasherInterface $userPasswordHasher)
    {
        $this->userManager = $userManager;
        $this->userPasswordHasher = $userPasswordHasher;
    }

    /**
     * @param FormInterface $form
     * @return User
     */
    public function processEditForm(FormInterface $form): User
    {
        $plainPassword = $form->get('plainPassword')->getData();
        $newEmail = $form->get('newEmail')->getData();

        /** @var User $user */
        $user = $form->getData();

        if ($newEmail) {
            $user->setEmail($newEmail);
        }

        if ($plainPassword) {
            $user->setPassword($this->userPasswordHasher->hashPassword($user, $plainPassword));
        }

        $this->userManager->save($user);

        return $user;
    }
}