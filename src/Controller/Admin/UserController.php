<?php

namespace App\Controller\Admin;

use App\Entity\StaticStorage\Role;
use App\Entity\User;
use App\Form\Admin\EditUserFormType;
use App\Form\Handler\UserFormHandler;
use App\Repository\UserRepository;
use App\Utils\Manager\UserManager;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Annotation\Route;

#[Route('/admin/user', name: 'admin_user_')]
class UserController extends AbstractController
{
    #[Route('/list', name: 'list')]
    public function list(UserRepository $userRepository): Response
    {
        if (!$this->isGranted(Role::SUPER_ADMIN)) {
            return $this->redirectToRoute('admin_dashboard_show');
        }

        return $this->render('admin/user/list.html.twig', [
            'users' => $userRepository->findBy(['isDeleted' => false], ['id' => 'DESC']),
        ]);
    }

    #[Route('/edit/{id}', name: 'edit')]
    #[Route('/add', name: 'add')]
    public function edit(
        Request         $request,
        UserFormHandler $userFormHandler,
        User            $user = null,
    ): Response
    {
        if (!$this->isGranted(Role::SUPER_ADMIN)) {
            return $this->redirectToRoute('admin_dashboard_show');
        }

        if (!$user) {
            $user = new User();
        }

        $form = $this->createForm(EditUserFormType::class, $user);
        $form->handleRequest($request);

        if ($form->isSubmitted() && $form->isValid()) {

            $user = $userFormHandler->processEditForm($form);
            $this->addFlash('success', 'The changes have been saved');

            return $this->redirectToRoute('admin_user_edit', ['id' => $user->getId()]);
        }

        if ($form->isSubmitted() && !$form->isValid()) {
            $this->addFlash('warning', 'Something went wrong');
        }

        return $this->render('admin/user/edit.html.twig', [
            'user' => $user,
            'form' => $form->createView(),
        ]);
    }

    #[Route('/delete/{id}', name: 'delete')]
    public function delete(UserManager $userManager, User $user): Response
    {
        if (!$this->isGranted(Role::SUPER_ADMIN)) {
            return $this->redirectToRoute('admin_dashboard_show');
        }

        $userManager->remove($user);

        $this->addFlash('warning', 'Пользователь ' . $user->getEmail() . ' удален!');

        return $this->redirectToRoute('admin_user_list');
    }
}
