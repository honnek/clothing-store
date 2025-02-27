<?php

namespace App\Form\Admin\FilterType;

use App\Entity\StaticStorage\OrderStatus;
use App\Entity\User;
use App\Form\DTO\EditOrderModel;
use Lexik\Bundle\FormFilterBundle\Filter\FilterOperands;
use Lexik\Bundle\FormFilterBundle\Filter\Form\Type as Filters;
use Symfony\Component\Form\AbstractType;
use Symfony\Component\Form\FormBuilderInterface;
use Symfony\Component\OptionsResolver\OptionsResolver;

class OrderFilterForm extends AbstractType
{
    public function buildForm(FormBuilderInterface $builder, array $options): void
    {
        $builder
            ->add('id', Filters\NumberFilterType::class, [
                'label' => 'id',
                'attr' => [
                    'class' => 'form-control'
                ]
            ])
            ->add('owner', Filters\EntityFilterType::class, [
                'label' => 'owner',
                'class' => User::class,
                'choice_label' => function($user) {
                    return sprintf('#%s %s', $user->getId(), $user->getEmail());
                },
                'attr' => [
                    'class' => 'form-control'
                ]
            ])
            ->add('status', Filters\ChoiceFilterType::class, [
                'label' => 'status',
                'choices' => array_flip(OrderStatus::getList()),
                'attr' => [
                    'class' => 'form-control'
                ]
            ])
            ->add('totalPrice', Filters\NumberRangeFilterType::class, [
                'label' => 'total price',
                'left_number_options' => [
                    'label' => 'From',
                    'condition_operator' => FilterOperands::OPERATOR_GREATER_THAN_EQUAL,
                    'attr' => [
                        'class' => 'form-control'
                    ]
                ],
                'right_number_options' => [
                    'label' => 'to',
                    'condition_operator' => FilterOperands::OPERATOR_LOWER_THAN_EQUAL,
                    'attr' => [
                        'class' => 'form-control'
                    ]
                ],
            ])
            ->add('createdAt', Filters\DateTimeRangeFilterType::class, [
                'label' => 'Created at',
                'left_datetime_options' => [
                    'label' => 'From',
                    'widget' => 'single_text',
                    'attr' => [
                        'class' => 'form-control'
                    ]
                ],
                'right_datetime_options' => [
                    'label' => 'To',
                    'widget' => 'single_text',
                    'attr' => [
                        'class' => 'form-control'
                    ]
                ],
            ])
        ;
    }

    public function getBlockPrefix(): string
    {
        return 'oder_filter_form';
    }
    
    public function configureOptions(OptionsResolver $resolver): void
    {
        $resolver->setDefaults([
            'data_class' => EditOrderModel::class,
            'method' => 'GET',
            'validation_groups' => ['filtering']
        ]);
    }
}
