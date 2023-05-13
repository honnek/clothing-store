<?php

namespace App\Form\Main;

use App\Entity\User;
use Symfony\Component\Form\AbstractType;
use Symfony\Component\Form\Extension\Core\Type\IntegerType;
use Symfony\Component\Form\FormBuilderInterface;
use Symfony\Component\OptionsResolver\OptionsResolver;

class ProfileEditFormType extends AbstractType
{
    public function buildForm(FormBuilderInterface $builder, array $options): void
    {
        $builder
            ->add('fullName', \Symfony\Component\Form\Extension\Core\Type\TextType::class, [
                'label' => 'Enter your full name'
            ])
            ->add('phone', \Symfony\Component\Form\Extension\Core\Type\TextType::class, [
                'label' => 'Enter your phone'
            ])
            ->add('address', \Symfony\Component\Form\Extension\Core\Type\TextType::class, [
                'label' => 'Enter your address'
            ])
            ->add('zipcode', IntegerType::class, [
                'label' => 'Enter your zipcode'
            ]);
    }

    public function configureOptions(OptionsResolver $resolver): void
    {
        $resolver->setDefaults([
            'data_class' => User::class,
        ]);
    }
}
