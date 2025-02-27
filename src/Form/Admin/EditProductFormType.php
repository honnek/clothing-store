<?php

namespace App\Form\Admin;

use App\Entity\Category;
use App\Form\DTO\EditProductModel;
use Symfony\Bridge\Doctrine\Form\Type\EntityType;
use Symfony\Component\Form\AbstractType;
use Symfony\Component\Form\Extension\Core\Type\CheckboxType;
use Symfony\Component\Form\Extension\Core\Type\FileType;
use Symfony\Component\Form\Extension\Core\Type\IntegerType;
use Symfony\Component\Form\Extension\Core\Type\NumberType;
use Symfony\Component\Form\Extension\Core\Type\TextareaType;
use Symfony\Component\Form\Extension\Core\Type\TextType;
use Symfony\Component\Form\FormBuilderInterface;
use Symfony\Component\OptionsResolver\OptionsResolver;

/**
 *
 */
class EditProductFormType extends AbstractType
{
    public function buildForm(FormBuilderInterface $builder, array $options): void
    {
        $builder
            ->add('title', TextType::class, [
                'label' => 'Title (from class)',
                'required' => true,
            ])
            ->add('price', NumberType::class, [
                'label' => 'Price (from class)',
                'scale' => 2,
                'html5' => true,
                'attr' => [
                    'step' => 0.01,
                ]
            ])
            ->add('quality', IntegerType::class, [
                'label' => 'quality',
                'attr' => [
                    'class' => 'form-control',
                ]
            ])
            ->add('description', TextareaType::class, [
                'label' => 'description',
                'required' => true,
                'attr' => [
                    'class' => 'form-control-file',
                ]
            ])
            ->add('category', EntityType::class, [
                'label' => 'category',
                'required' => true,
                'class' => Category::class,
                'choice_label' => 'title',
                'attr' => [
                    'class' => 'form-control',
                ]
            ])
            ->add('newImage', FileType::class, [
                'label' => 'Chose new image',
                'required' => false,
                'attr' => [
                    'class' => 'form-check-input',
                ],
            ])
            ->add('isPublished', CheckboxType::class, [
                'label' => 'Is published',
                'required' => false,

                'attr' => [
                    'class' => 'form-check-input',
                ],
            ])
            ->add('isDeleted', CheckboxType::class, [
                'label' => 'Is deleted',
                'required' => false,
                'attr' => [
                    'class' => 'form-check-input',
                ],
            ]);
    }

    public function configureOptions(OptionsResolver $resolver): void
    {
        $resolver->setDefaults([
            'data_class' => EditProductModel::class,
        ]);
    }
}
