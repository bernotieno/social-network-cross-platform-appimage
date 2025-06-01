'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useAuth } from '@/hooks/useAuth';
import Input from '@/components/Input';
import Button from '@/components/Button';
import styles from '@/styles/Auth.module.css';

export default function Register() {
  const [formData, setFormData] = useState({
    username: '',
    email: '',
    password: '',
    confirmPassword: '',
    firstName: '',
    lastName: '',
    dateOfBirth: '',
    bio: '',
  });
  const [avatar, setAvatar] = useState(null);
  const [avatarPreview, setAvatarPreview] = useState(null);
  const [errors, setErrors] = useState({});
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [registerError, setRegisterError] = useState('');

  const { register } = useAuth();
  const router = useRouter();

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }));

    // Clear error when user types
    if (errors[name]) {
      setErrors((prev) => ({
        ...prev,
        [name]: '',
      }));
    }

    // Clear register error when user types
    if (registerError) {
      setRegisterError('');
    }
  };

  const handleAvatarChange = (e) => {
    const file = e.target.files[0];
    if (file) {
      setAvatar(file);
      const reader = new FileReader();
      reader.onloadend = () => {
        setAvatarPreview(reader.result);
      };
      reader.readAsDataURL(file);
    }
  };

  const validateForm = () => {
    const newErrors = {};

    if (!formData.username.trim()) {
      newErrors.username = 'Username is required';
    } else if (formData.username.length < 3) {
      newErrors.username = 'Username must be at least 3 characters';
    }

    if (!formData.email.trim()) {
      newErrors.email = 'Email is required';
    } else if (!/\S+@\S+\.\S+/.test(formData.email)) {
      newErrors.email = 'Email is invalid';
    }

    if (!formData.password) {
      newErrors.password = 'Password is required';
    } else if (formData.password.length < 6) {
      newErrors.password = 'Password must be at least 6 characters';
    }

    if (formData.password !== formData.confirmPassword) {
      newErrors.confirmPassword = 'Passwords do not match';
    }

    if (!formData.firstName.trim()) {
      newErrors.firstName = 'First name is required';
    }

    if (!formData.lastName.trim()) {
      newErrors.lastName = 'Last name is required';
    }

    if (!formData.dateOfBirth) {
      newErrors.dateOfBirth = 'Date of birth is required';
    } else {
      // Validate age (must be at least 13 years old)
      const birthDate = new Date(formData.dateOfBirth);
      const today = new Date();
      let age = today.getFullYear() - birthDate.getFullYear();
      const monthDiff = today.getMonth() - birthDate.getMonth();

      if (monthDiff < 0 || (monthDiff === 0 && today.getDate() < birthDate.getDate())) {
        age--;
      }

      if (age < 13) {
        newErrors.dateOfBirth = 'You must be at least 13 years old to register';
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    setIsSubmitting(true);

    try {
      // Create FormData for multipart/form-data submission
      const formDataToSend = new FormData();

      // Add all form fields except confirmPassword
      Object.keys(formData).forEach(key => {
        if (key !== 'confirmPassword' && formData[key]) {
          formDataToSend.append(key, formData[key]);
        }
      });

      // Generate fullName from firstName and lastName
      const fullName = `${formData.firstName} ${formData.lastName}`.trim();
      formDataToSend.set('fullName', fullName);

      // Add avatar file if selected
      if (avatar) {
        formDataToSend.append('avatar', avatar);
      }

      const result = await register(formDataToSend);

      if (result.success) {
        router.push('/');
      } else {
        setRegisterError(result.error);
      }
    } catch (error) {
      setRegisterError('An unexpected error occurred. Please try again.');
      console.error('Registration error:', error);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className={styles.authContainer}>
      <div className={styles.authCard}>
        <h1 className={styles.authTitle}>Create Account</h1>

        {registerError && (
          <div className={styles.errorAlert}>
            {registerError}
          </div>
        )}

        <form onSubmit={handleSubmit} className={styles.authForm}>
          <Input
            label="Username/Nickname"
            type="text"
            id="username"
            name="username"
            value={formData.username}
            onChange={handleChange}
            placeholder="Choose a username/nickname"
            error={errors.username}
            required
            fullWidth
          />

          <div className={styles.nameRow}>
            <Input
              label="First Name"
              type="text"
              id="firstName"
              name="firstName"
              value={formData.firstName}
              onChange={handleChange}
              placeholder="First name"
              error={errors.firstName}
              required
              fullWidth
            />

            <Input
              label="Last Name"
              type="text"
              id="lastName"
              name="lastName"
              value={formData.lastName}
              onChange={handleChange}
              placeholder="Last name"
              error={errors.lastName}
              required
              fullWidth
            />
          </div>

          <Input
            label="Email"
            type="email"
            id="email"
            name="email"
            value={formData.email}
            onChange={handleChange}
            placeholder="Enter your email"
            error={errors.email}
            required
            fullWidth
          />

          <Input
            label="Date of Birth"
            type="date"
            id="dateOfBirth"
            name="dateOfBirth"
            value={formData.dateOfBirth}
            onChange={handleChange}
            error={errors.dateOfBirth}
            required
            fullWidth
          />

          <Input
            label="Password"
            type="password"
            id="password"
            name="password"
            value={formData.password}
            onChange={handleChange}
            placeholder="Create a password"
            error={errors.password}
            required
            fullWidth
          />

          <Input
            label="Confirm Password"
            type="password"
            id="confirmPassword"
            name="confirmPassword"
            value={formData.confirmPassword}
            onChange={handleChange}
            placeholder="Confirm your password"
            error={errors.confirmPassword}
            required
            fullWidth
          />

          <div className={styles.formGroup}>
            <label htmlFor="bio" className={styles.label}>
              Bio (Optional)
            </label>
            <textarea
              id="bio"
              name="bio"
              value={formData.bio}
              onChange={handleChange}
              placeholder="Tell us about yourself..."
              className={styles.textarea}
              rows="3"
            />
            {errors.bio && <p className={styles.errorMessage}>{errors.bio}</p>}
          </div>

          <div className={styles.formGroup}>
            <label htmlFor="avatar" className={styles.label}>
              Avatar/Profile Picture (Optional)
            </label>
            <input
              type="file"
              id="avatar"
              accept="image/*"
              onChange={handleAvatarChange}
              className={styles.fileInput}
            />
            {avatarPreview && (
              <div className={styles.avatarPreview}>
                <img src={avatarPreview} alt="Avatar preview" className={styles.previewImage} />
              </div>
            )}
            {errors.avatar && <p className={styles.errorMessage}>{errors.avatar}</p>}
          </div>

          <Button
            type="submit"
            variant="primary"
            size="large"
            fullWidth
            disabled={isSubmitting}
          >
            {isSubmitting ? 'Creating Account...' : 'Register'}
          </Button>
        </form>

        <div className={styles.authLinks}>
          <p>
            Already have an account?{' '}
            <Link href="/auth/login" className={styles.authLink}>
              Login
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}